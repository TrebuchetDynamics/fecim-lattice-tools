package server

import (
	"context"
	"errors"
	"io"
	"os"
	"path"
	"sort"
	"strings"
	"sync"

	"fecim-lattice-tools/workbench/edaexport"
	"fecim-lattice-tools/workbench/experiment"
	"fecim-lattice-tools/workbench/moduleapi"
	"fecim-lattice-tools/workbench/moduleapi/httpapi"
	"fecim-lattice-tools/workbench/presentation"
	"fecim-lattice-tools/workbench/project"
	"fecim-lattice-tools/workbench/study"
)

const maxProjects = 32
const maxModuleSessions = 16
const discoveryLimit = 1000

type openedProject struct {
	id, grant, path string
	root            *os.Root
	info            os.FileInfo
	editor          *project.ConfigurationEditor
	gate            chan struct{}
	life            sync.RWMutex
	config          sync.RWMutex
	closed          bool
}

func (s *Server) closeProjects() {
	s.mu.Lock()
	projects := s.projects
	s.projects = map[string]*openedProject{}
	s.mu.Unlock()
	for _, p := range projects {
		p.life.Lock()
		p.closed = true
		_ = p.root.Close()
		p.life.Unlock()
	}
}
func (s *Server) borrow(id string, mutation bool) (*openedProject, func(), error) {
	s.mu.Lock()
	p := s.projects[id]
	ended := s.closed || s.draining
	s.mu.Unlock()
	if p == nil {
		return nil, nil, httpapi.Fail(404, "Project handle not found")
	}
	if ended {
		return nil, nil, httpapi.Fail(503, "Session ended")
	}
	p.life.RLock()
	if p.closed {
		p.life.RUnlock()
		return nil, nil, httpapi.Fail(404, "Project handle closed")
	}
	if mutation {
		select {
		case p.gate <- struct{}{}:
		default:
			p.life.RUnlock()
			return nil, nil, httpapi.Fail(409, "A project operation is active")
		}
	}
	return p, func() {
		if mutation {
			<-p.gate
		}
		p.life.RUnlock()
	}, nil
}
func (s *Server) openProject(grant, relative string) (*openedProject, error) {
	if !safeRelative(relative) {
		return nil, httpapi.Fail(400, "Choose a safe relative project directory")
	}
	root := s.roots[grant]
	if root == nil {
		return nil, httpapi.Fail(400, "Unknown granted root")
	}
	// Reject symlink components to avoid overlapping project aliases. Each opened
	// child is rooted and pinned, even if a directory is renamed externally.
	current := root
	owned := false
	for _, part := range strings.Split(relative, "/") {
		info, err := current.Lstat(part)
		if err != nil || !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
			if owned {
				current.Close()
			}
			return nil, httpapi.Fail(400, "Project directory unavailable within grant")
		}
		next, err := current.OpenRoot(part)
		if owned {
			current.Close()
		}
		if err != nil {
			return nil, httpapi.Fail(400, "Project directory unavailable within grant")
		}
		actual, statErr := next.Stat(".")
		if statErr != nil || !os.SameFile(info, actual) {
			next.Close()
			return nil, httpapi.Fail(409, "Project directory changed during admission")
		}
		current = next
		owned = true
	}
	child := current
	info, err := child.Stat(".")
	if err != nil {
		child.Close()
		return nil, httpapi.Fail(400, "Project directory unavailable")
	}
	// PrepareConfiguration with zero edits uses bounded snapshot decoding and the
	// same whole-bundle validator; LoadRoot now also preflights before expansion.
	if _, _, err := study.ValidateProjectRoot(child, nil); err != nil {
		child.Close()
		return nil, httpapi.Fail(400, "Project bundle is invalid or unavailable")
	}
	editor, err := project.NewConfigurationEditor(child, nil)
	if err != nil {
		child.Close()
		return nil, httpapi.Fail(400, "Project editor unavailable")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed || s.draining {
		child.Close()
		return nil, httpapi.Fail(503, "Session ended")
	}
	for _, p := range s.projects {
		if os.SameFile(p.info, info) {
			child.Close()
			return p, nil
		}
		if p.grant == grant && (strings.HasPrefix(relative, p.path+"/") || strings.HasPrefix(p.path, relative+"/")) {
			child.Close()
			return nil, httpapi.Fail(409, "Overlapping project handles are not allowed")
		}
	}
	if len(s.projects) >= maxProjects {
		child.Close()
		return nil, httpapi.Fail(503, "Project handle capacity reached; close a project")
	}
	p := &openedProject{id: randomSecret(), grant: grant, path: relative, root: child, info: info, editor: editor, gate: make(chan struct{}, 1)}
	s.projects[p.id] = p
	return p, nil
}
func configuration(p *openedProject) (httpapi.Configuration, error) {
	snapshot, err := project.ReadConfiguration(p.root)
	if err != nil {
		return httpapi.Configuration{}, httpapi.Fail(400, "Configuration is unavailable")
	}
	fields, err := snapshot.Fields()
	if err != nil {
		return httpapi.Configuration{}, httpapi.Fail(400, "Configuration fields are unavailable")
	}
	bundle, err := project.LoadRoot(p.root, nil)
	if err != nil {
		return httpapi.Configuration{}, httpapi.Fail(400, "Project bundle is invalid or unavailable")
	}
	presented := make([]moduleapi.ConfigurationField, 0, len(fields))
	for _, f := range fields {
		presented = append(presented, moduleapi.FieldFromConfiguration(f, moduleapi.ConfigurationUnits(f, bundle)))
	}
	return httpapi.Configuration{ID: p.id, Revision: snapshot.Revision(), Fields: presented, Evidence: "simulation-default; published parameters are not device measurements"}, nil
}
func projectView(p *openedProject) (httpapi.Project, error) {
	c, err := configuration(p)
	if err != nil {
		return httpapi.Project{}, err
	}
	bundle, err := project.LoadRoot(p.root, nil)
	if err != nil {
		return httpapi.Project{}, httpapi.Fail(400, "Project bundle is invalid or unavailable")
	}
	return httpapi.Project{Configuration: c, Name: bundle.Project.Name, ModelVersion: bundle.Project.ModelVersion, Citations: bundle.Project.Citations, Inputs: bundle.Inputs}, nil
}
func result(value any) httpapi.Result { return httpapi.Result{Value: value} }
func projectResult(p *openedProject) (httpapi.Result, error) {
	v, e := projectView(p)
	return httpapi.Result{Value: v, ETag: v.Revision}, e
}
func (s *Server) Execute(ctx context.Context, op string, input any) (httpapi.Result, error) {
	releaseSession, err := s.beginSession(ctx)
	if err != nil {
		return httpapi.Result{}, err
	}
	defer releaseSession()
	switch op {
	case "bind-resource":
		return s.bindResource(input.(*httpapi.BindResourceInput))
	case "health":
		return result(httpapi.Health{Identity: s.manifest.Identity, UI: s.manifest.UI, Protocol: moduleapi.SchemaVersion, Limits: map[string]int{"body_bytes": MaxBodyBytes, "projects": maxProjects, "jobs": maxJobs, "streams": cap(s.streams), "workers": 4, "module_sessions": maxModuleSessions, "resource_bytes": int(maxResourceBytes), "resources": maxResources}}), nil
	case "modules":
		return result(httpapi.Modules{Modules: s.moduleDescriptors}), nil
	case "projects":
		return s.listProjects(input.(*httpapi.ListProjectsInput))
	case "create-project":
		q := input.(*httpapi.CreateProjectInput).Body
		root := s.roots[q.Root]
		if root == nil || !safeRelative(q.Name) || strings.Contains(q.Name, "/") {
			return httpapi.Result{}, httpapi.Fail(400, "Choose a granted root and one safe directory name")
		}
		opts := study.DefaultStarterSweepOptions()
		if q.Options != nil {
			opts = *q.Options
		}
		// Reserve creation capacity under a short metadata lock; filesystem work is
		// not under the global lock. Mkdir is exclusive across simultaneous creates.
		s.mu.Lock()
		full := len(s.projects) >= maxProjects
		s.mu.Unlock()
		if full {
			return httpapi.Result{}, httpapi.Fail(503, "Project handle capacity reached")
		}
		if err := study.CreateStarterProjectRoot(root, q.Name, opts); err != nil {
			if errors.Is(err, os.ErrExist) {
				return httpapi.Result{}, httpapi.Fail(409, "Destination exists; choose a new name")
			}
			return httpapi.Result{}, httpapi.Fail(400, "Starter options or destination are unavailable")
		}
		p, e := s.openProject(q.Root, q.Name)
		if e != nil {
			return httpapi.Result{}, e
		}
		p.life.RLock()
		defer p.life.RUnlock()
		p.config.RLock()
		defer p.config.RUnlock()
		if p.closed {
			return httpapi.Result{}, httpapi.Fail(404, "Project closed")
		}
		return projectResult(p)
	case "open-project":
		q := input.(*httpapi.OpenProjectInput).Body
		p, e := s.openProject(q.Root, q.Path)
		if e != nil {
			return httpapi.Result{}, e
		}
		p.life.RLock()
		defer p.life.RUnlock()
		p.config.RLock()
		defer p.config.RUnlock()
		if p.closed {
			return httpapi.Result{}, httpapi.Fail(404, "Project closed")
		}
		return projectResult(p)
	case "execution":
		return s.execution(input.(*httpapi.ExecutionInput).Execution, false)
	case "cancel":
		return s.execution(input.(*httpapi.CancelInput).Execution, true)
	case "start":
		return s.start(ctx, input.(*httpapi.StartInput))
	case "active":
		q := input.(*httpapi.ActiveInput)
		s.mu.Lock()
		defer s.mu.Unlock()
		if s.projects[q.Project] == nil {
			return httpapi.Result{}, httpapi.Fail(404, "Project handle not found")
		}
		out := httpapi.Jobs{Experiments: []moduleapi.Job{}}
		for _, id := range s.jobOrder {
			j := s.jobs[id]
			if j == nil {
				continue
			}
			v := j.snapshot()
			if v.Project == q.Project && (q.State != "active" || activeState(v.State)) {
				out.Experiments = append(out.Experiments, v)
			}
		}
		return result(out), nil
	case "module-create", "module-snapshot", "module-action", "module-close", "module-start", "module-job", "module-cancel":
		return s.moduleExecute(ctx, op, input)
	}
	var id string
	switch q := input.(type) {
	case *httpapi.ProjectInput:
		id = q.Project
	case *httpapi.ProjectMutationInput:
		id = q.Project
	case *httpapi.PatchInput:
		id = q.Project
	case *httpapi.RunsInput:
		id = q.Project
	case *httpapi.AnalysisInput:
		id = q.Project
	case *httpapi.RunInput:
		id = q.Project
	case *httpapi.ExportInput:
		id = q.Project
	case *httpapi.ReportInput:
		id = q.Project
	case *httpapi.DownloadExportInput:
		id = q.Project
	}
	mutation := op == "patch" || op == "validate" || op == "verify" || op == "reports" || op == "export" || op == "close-project"
	if op == "close-project" {
		return s.closeProject(id)
	}
	p, release, err := s.borrow(id, mutation)
	if err != nil {
		return httpapi.Result{}, err
	}
	defer release()
	if op == "patch" {
		p.config.Lock()
		defer p.config.Unlock()
	} else {
		p.config.RLock()
		defer p.config.RUnlock()
	}
	switch op {
	case "project":
		return projectResult(p)
	case "configuration":
		v, e := configuration(p)
		return httpapi.Result{Value: v, ETag: v.Revision}, e
	case "patch":
		q := input.(*httpapi.PatchInput)
		changes := make([]project.FieldChange, len(q.Body.Changes))
		for i, c := range q.Body.Changes {
			changes[i] = project.FieldChange{Path: c.Path, Value: c.Value}
		}
		_, err := p.editor.Save(ctx, q.IfMatch, q.Body.Document, changes)
		if err != nil {
			if errors.Is(err, project.ErrRevisionRequired) {
				return httpapi.Result{}, httpapi.Fail(428, "Exact If-Match configuration revision required")
			}
			var conflict *project.ConfigurationConflict
			if errors.As(err, &conflict) {
				problem := httpapi.Fail(412, "Configuration changed; review literal input text before retrying")
				problem.Revision = conflict.CurrentRevision
				current, _ := configuration(p)
				for _, f := range conflict.Fields {
					metadata := moduleapi.FieldFromConfiguration(project.ConfigurationField{Document: q.Body.Document, Path: f.Path, Kind: "string"}, "")
					for _, field := range current.Fields {
						if field.Document == q.Body.Document && field.Path == f.Path {
							metadata = field
							break
						}
					}
					yoursField := metadata.ConfigurationField
					yoursField.Value = f.Yours
					diskField := metadata.ConfigurationField
					diskField.Value = f.OnDisk
					yours := moduleapi.FieldFromConfiguration(yoursField, metadata.Unit).Presentation
					onDisk := moduleapi.FieldFromConfiguration(diskField, metadata.Unit).Presentation
					if !f.OnDiskAvailable {
						onDisk = presentation.Value{Display: "Not available", Speakable: "On-disk field is unavailable", Availability: "unavailable", Reason: "On-disk field is unavailable"}
					}
					problem.Fields = append(problem.Fields, httpapi.FieldConflict{Path: f.Path, Label: metadata.Label, Unit: metadata.Unit, Yours: f.Yours, OnDisk: f.OnDisk, OnDiskAvailable: f.OnDiskAvailable, YoursPresentation: yours, OnDiskPresentation: onDisk, Reason: "Literal editor text; no unit conversion or automatic merge", Representation: "literal-text"})
				}
				return httpapi.Result{}, problem
			}
			var issue *project.ConfigurationFieldIssue
			if errors.As(err, &issue) {
				label := moduleapi.FieldFromConfiguration(project.ConfigurationField{Document: issue.Document, Path: issue.Path}, "").Label
				if current, e := configuration(p); e == nil {
					for _, field := range current.Fields {
						if field.Document == issue.Document && field.Path == issue.Path {
							label = field.Label
							break
						}
					}
				}
				p := httpapi.Fail(400, "Field input rejected; review the indicated field")
				p.Issues = []httpapi.FieldIssue{{Document: issue.Document, Path: issue.Path, Label: label, Reason: issue.Reason}}
				return httpapi.Result{}, p
			}
			return httpapi.Result{}, httpapi.Fail(400, "Field patch rejected; check document, field types, ranges and whole-bundle validity")
		}
		v, e := configuration(p)
		return httpapi.Result{Value: v, ETag: v.Revision}, e
	case "validate":
		_, n, e := study.ValidateProjectRoot(p.root, nil)
		if e != nil {
			return httpapi.Result{}, httpapi.Fail(400, "Project bundle is invalid or unavailable")
		}
		return result(httpapi.Validation{Points: n, Valid: true}), nil
	case "run":
		runID := input.(*httpapi.RunInput).Run
		if !scientificID(runID) {
			return httpapi.Result{}, httpapi.Fail(400, "Invalid scientific run ID")
		}
		record, ok, e := experiment.LoadRunRecordRoot(p.root, runID)
		if e != nil {
			return httpapi.Result{}, httpapi.Fail(409, "Stored run failed integrity review")
		}
		if !ok {
			return httpapi.Result{}, httpapi.Fail(404, "Stored run not found")
		}
		return result(record), nil
	case "runs", "analysis", "verify":
		a, e := readAnalysis(p)
		if e != nil {
			return httpapi.Result{}, e
		}
		query := study.ResultsQuery{}
		var compare string
		switch q := input.(type) {
		case *httpapi.RunsInput:
			query = q.Query()
		case *httpapi.AnalysisInput:
			query = q.Query()
			compare = q.Compare
		}
		page, e := study.QueryResults(a, query)
		if e != nil {
			return httpapi.Result{}, httpapi.Fail(400, "Unsupported results filter, sort or page")
		}
		if op == "runs" {
			return result(page), nil
		}
		plot, e := study.PlotResults(a, query)
		if e != nil {
			return httpapi.Result{}, httpapi.Fail(400, "Results plot unavailable for this query")
		}
		out := httpapi.Analysis{Counts: a.Counts, Results: page, Plot: plot}
		if compare != "" {
			comparison, e := study.CompareResults(a, strings.Split(compare, ","))
			if e != nil {
				return httpapi.Result{}, httpapi.Fail(400, "Select two to four distinct stored run IDs")
			}
			out.Comparison = &comparison
		}
		return result(out), nil
	case "reports":
		if _, e := study.GenerateReportRoot(p.root, nil); e != nil {
			return httpapi.Result{}, httpapi.Fail(409, "Reports unavailable; review stored runs and project validity")
		}
		return result(httpapi.Artifacts{Files: reportFiles}), nil
	case "export":
		id := input.(*httpapi.ExportInput).Body.RunID
		if !scientificID(id) {
			return httpapi.Result{}, httpapi.Fail(400, "Invalid scientific run ID")
		}
		if _, e := study.ExportRunRoot(p.root, id); e != nil {
			return httpapi.Result{}, httpapi.Fail(409, "Export unavailable; select an eligible verified run and unused export destination")
		}
		return result(httpapi.Artifacts{RunID: id, Files: exportFiles}), nil
	case "report-download":
		return downloadFile(p, "reports", input.(*httpapi.ReportInput).File, reportFiles)
	case "export-download":
		q := input.(*httpapi.DownloadExportInput)
		if !scientificID(q.Run) {
			return httpapi.Result{}, httpapi.Fail(400, "Invalid scientific run ID")
		}
		if !edaexport.AllowedArtifact(q.Artifact) {
			return httpapi.Result{}, httpapi.Fail(404, "Artifact not enumerated")
		}
		data, err := edaexport.ReadArtifactRoot(p.root, q.Run, q.Artifact)
		if err != nil {
			return httpapi.Result{}, httpapi.Fail(409, "Export unavailable or failed integrity review")
		}
		media := "text/plain; charset=utf-8"
		if strings.HasSuffix(q.Artifact, ".json") {
			media = "application/json"
		}
		if q.Artifact == "package.zip" {
			media = "application/zip"
		}
		return result(httpapi.Download{Data: data, Type: media, Name: q.Artifact}), nil
	}
	return httpapi.Result{}, httpapi.Fail(404, "Operation not found")
}
func scientificID(s string) bool {
	if len(s) != 64 {
		return false
	}
	for _, c := range s {
		if !(c >= '0' && c <= '9' || c >= 'a' && c <= 'f') {
			return false
		}
	}
	return true
}
func readAnalysis(p *openedProject) (experiment.Analysis, error) {
	bundle, err := project.LoadRoot(p.root, nil)
	if err != nil {
		return experiment.Analysis{}, httpapi.Fail(400, "Project bundle is invalid or unavailable")
	}
	runs, err := experiment.LoadRunsRoot(p.root)
	if err != nil {
		return experiment.Analysis{}, httpapi.Fail(409, "Stored runs failed integrity review")
	}
	return experiment.Analyze(bundle, runs), nil
}
func (s *Server) closeProject(id string) (httpapi.Result, error) {
	p, _, err := s.borrow(id, true)
	if err != nil {
		return httpapi.Result{}, err
	}
	// Mark/revoke while holding the operation slot. Upgrade lifetime only after
	// dropping this borrow, and let already-admitted reads finish first.
	s.mu.Lock()
	delete(s.projects, id)
	s.mu.Unlock()
	p.life.RUnlock()
	p.life.Lock()
	p.closed = true
	_ = p.root.Close()
	p.life.Unlock()
	<-p.gate
	return result(httpapi.Status{Status: "closed; scientific files retained"}), nil
}
func (s *Server) listProjects(q *httpapi.ListProjectsInput) (httpapi.Result, error) {
	page, size := q.Page, q.PageSize
	if page == 0 {
		page = 1
	}
	if size == 0 {
		size = 32
	}
	candidates := []httpapi.ProjectCandidate{}
	grants := make([]string, 0, len(s.roots))
	for id := range s.roots {
		grants = append(grants, id)
	}
	sort.Strings(grants)
	scanned := 0
	for _, id := range grants {
		dir, err := s.roots[id].Open(".")
		if err != nil {
			continue
		}
		entries, err := dir.ReadDir(discoveryLimit - scanned + 1)
		dir.Close()
		if err != nil && !errors.Is(err, io.EOF) {
			continue
		}
		sort.Slice(entries, func(i, j int) bool { return entries[i].Name() < entries[j].Name() })
		for _, entry := range entries {
			scanned++
			if scanned > discoveryLimit {
				break
			}
			if !entry.IsDir() || !safeRelative(entry.Name()) {
				continue
			}
			child, e := s.roots[id].OpenRoot(entry.Name())
			if e != nil {
				continue
			}
			_, e = child.Stat("project.yaml")
			child.Close()
			if e != nil {
				continue
			}
			candidate := httpapi.ProjectCandidate{Root: id, Path: entry.Name()}
			s.mu.Lock()
			for _, p := range s.projects {
				if p.grant == id && p.path == entry.Name() {
					candidate.OpenID = p.id
				}
			}
			s.mu.Unlock()
			candidates = append(candidates, candidate)
		}
		if scanned > discoveryLimit {
			break
		}
	}
	start := len(candidates)
	if page <= 1+len(candidates)/size {
		start = (page - 1) * size
		if start > len(candidates) {
			start = len(candidates)
		}
	}
	end := start + size
	if end > len(candidates) {
		end = len(candidates)
	}
	return result(httpapi.Projects{Projects: candidates[start:end], Page: page, PageSize: size, HasMore: end < len(candidates)}), nil
}

var reportFiles = []string{"results.json", "results.csv", "report.md"}
var exportFiles = []string{"design.spice", "design.v", "design.def", "cell.lef", "cell.lib", "manifest.json"}

func downloadFile(p *openedProject, dir, name string, allowed []string) (httpapi.Result, error) {
	ok := false
	for _, f := range allowed {
		ok = ok || f == name
	}
	if !ok {
		return httpapi.Result{}, httpapi.Fail(404, "Artifact not enumerated")
	}
	parent, err := p.root.OpenRoot(dir)
	if err != nil {
		return httpapi.Result{}, httpapi.Fail(404, "Artifact unavailable")
	}
	defer parent.Close()
	info, err := parent.Lstat(name)
	if err != nil || !info.Mode().IsRegular() || info.Size() > 32<<20 {
		return httpapi.Result{}, httpapi.Fail(404, "Artifact unavailable or exceeds download limit")
	}
	file, err := parent.Open(name)
	if err != nil {
		return httpapi.Result{}, httpapi.Fail(404, "Artifact unavailable")
	}
	defer file.Close()
	after, err := file.Stat()
	if err != nil || !os.SameFile(info, after) || !after.Mode().IsRegular() {
		return httpapi.Result{}, httpapi.Fail(409, "Artifact changed during read")
	}
	data, err := io.ReadAll(io.LimitReader(file, 32<<20+1))
	if err != nil || len(data) > 32<<20 {
		return httpapi.Result{}, httpapi.Fail(413, "Artifact exceeds download limit")
	}
	media := "text/plain; charset=utf-8"
	switch path.Ext(name) {
	case ".json":
		media = "application/json"
	case ".csv":
		media = "text/csv; charset=utf-8"
	case ".md":
		media = "text/markdown; charset=utf-8"
	}
	return result(httpapi.Download{Data: data, Type: media, Name: name}), nil
}
