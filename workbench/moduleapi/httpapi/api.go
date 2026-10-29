// Package httpapi is the native-only Huma/stdlib transport. Do not import it in
// Worker entrypoints; moduleapi itself has no router/framework dependency.
package httpapi

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"reflect"
	"strconv"
	"strings"

	"fecim-lattice-tools/workbench/experiment"
	"fecim-lattice-tools/workbench/hysteresis"
	"fecim-lattice-tools/workbench/moduleapi"
	"fecim-lattice-tools/workbench/presentation"
	"fecim-lattice-tools/workbench/project"
	"fecim-lattice-tools/workbench/study"
	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humago"
)

type Problem struct {
	Type     string          `json:"type"`
	Title    string          `json:"title"`
	Status   int             `json:"status"`
	Detail   string          `json:"detail"`
	Fields   []FieldConflict `json:"fields,omitempty"`
	Issues   []FieldIssue    `json:"issues,omitempty"`
	Revision string          `json:"revision,omitempty"`
}

func (p *Problem) Error() string  { return p.Detail }
func (p *Problem) GetStatus() int { return p.Status }
func Fail(status int, detail string) *Problem {
	return &Problem{Type: "urn:fecim:problem:" + strconv.Itoa(status), Title: http.StatusText(status), Status: status, Detail: detail}
}

type FieldIssue struct {
	Document string `json:"document"`
	Path     string `json:"path"`
	Label    string `json:"label"`
	Reason   string `json:"reason"`
}

type FieldConflict struct {
	Path               string             `json:"path"`
	Label              string             `json:"label"`
	Yours              string             `json:"yours"`
	OnDisk             string             `json:"on_disk"`
	OnDiskAvailable    bool               `json:"on_disk_available"`
	Reason             string             `json:"reason,omitempty"`
	Representation     string             `json:"representation"` // literal input text, never a scientific numeric cell
	Unit               string             `json:"unit"`
	YoursPresentation  presentation.Value `json:"yours_presentation"`
	OnDiskPresentation presentation.Value `json:"on_disk_presentation"`
}
type Configuration struct {
	ID       string                         `json:"id"`
	Revision string                         `json:"revision"`
	Fields   []moduleapi.ConfigurationField `json:"fields"`
	Evidence string                         `json:"evidence"`
}
type Project struct {
	Configuration
	Name         string                  `json:"name"`
	ModelVersion string                  `json:"model_version"`
	Citations    []string                `json:"citations"`
	Inputs       []project.ResolvedInput `json:"inputs"`
}
type ProjectCandidate struct {
	Root   string `json:"root"`
	Path   string `json:"path"`
	OpenID string `json:"open_id,omitempty"`
}
type Projects struct {
	Projects []ProjectCandidate `json:"projects"`
	Page     int                `json:"page"`
	PageSize int                `json:"page_size"`
	HasMore  bool               `json:"has_more"`
}
type Validation struct {
	Points int  `json:"points"`
	Valid  bool `json:"valid"`
}
type Status struct {
	Status string `json:"status"`
}
type Jobs struct {
	Experiments []moduleapi.Job `json:"experiments"`
}
type Analysis struct {
	Plot       study.ResultsPlot        `json:"plot"`
	Counts     map[string]int           `json:"counts"`
	Results    study.ResultsPage        `json:"results"`
	Comparison *study.ResultsComparison `json:"comparison,omitempty"`
}
type Artifacts struct {
	Files []string `json:"files"`
	RunID string   `json:"run_id,omitempty"`
}
type Modules struct {
	Modules []moduleapi.ModuleDescriptor `json:"modules"`
}
type Health struct {
	Identity hysteresis.Identity `json:"identity"`
	UI       string              `json:"UI"`
	Protocol string              `json:"protocol"`
	Limits   map[string]int      `json:"limits"`
}
type BindResourceBody struct {
	Root string `json:"root"`
	Path string `json:"path"`
	Kind string `json:"kind"`
}
type BindResourceInput struct{ Body BindResourceBody }

type CreateProjectBody struct {
	Root    string                     `json:"root"`
	Name    string                     `json:"name"`
	Options *study.StarterSweepOptions `json:"options,omitempty"`
}
type OpenProjectBody struct {
	Root string `json:"root"`
	Path string `json:"path"`
}
type FieldChange struct {
	Path  string `json:"path" maxLength:"240"`
	Value string `json:"value" maxLength:"4096"`
}
type PatchBody struct {
	Document string        `json:"document" enum:"project.yaml,design.yaml,sweep.yaml"`
	Changes  []FieldChange `json:"changes" maxItems:"64"`
}
type StartBody struct {
	Workers int `json:"workers" minimum:"1" maximum:"4"`
}
type ExportBody struct {
	RunID string `json:"run_id"`
}
type Empty struct{}
type Nothing struct{}
type P struct {
	Project string `path:"project"`
}
type E struct {
	Execution string `path:"execution"`
}
type M struct {
	Module string `path:"module"`
}
type S struct {
	Session string `path:"session"`
}
type ListProjectsInput struct {
	Page     int `query:"page" minimum:"1"`
	PageSize int `query:"page_size" minimum:"1" maximum:"100"`
}
type CreateProjectInput struct{ Body CreateProjectBody }
type OpenProjectInput struct{ Body OpenProjectBody }
type ProjectInput struct{ P }
type ProjectMutationInput struct {
	P
	Body Empty
}
type PatchInput struct {
	P
	IfMatch string `header:"If-Match"`
	Body    PatchBody
}
type StartInput struct {
	P
	IfMatch        string `header:"If-Match"`
	IdempotencyKey string `header:"Idempotency-Key"`
	Body           StartBody
}
type ActiveInput struct {
	P
	State string `query:"state" enum:"active"`
}
type ExecutionInput struct{ E }
type CancelInput struct {
	E
	Body Empty
}
type RunsInput struct {
	P
	Page       int    `query:"page"`
	PageSize   int    `query:"page_size"`
	Filter     string `query:"filter"`
	Sort       string `query:"sort"`
	Descending bool   `query:"descending"`
	Revision   string `query:"revision"`
}

func (i *RunsInput) Query() study.ResultsQuery {
	return study.ResultsQuery{Page: i.Page, PageSize: i.PageSize, Filter: i.Filter, Sort: i.Sort, Descending: i.Descending, Revision: i.Revision}
}

type RunInput struct {
	P
	Run string `path:"run"`
}
type AnalysisInput struct {
	RunsInput
	Compare string `query:"compare"`
}
type ReportInput struct {
	P
	File string `path:"file"`
}
type ExportInput struct {
	P
	Body ExportBody
}
type DownloadExportInput struct {
	P
	Run      string `path:"run"`
	Artifact string `path:"artifact"`
}
type ModuleCreateInput struct {
	M
	Body Empty
}
type ModuleSessionInput struct{ S }
type ModuleActionInput struct {
	S
	Body moduleapi.Command
}
type ModuleCloseInput struct {
	S
	Body Empty
}
type ModuleJobInput struct {
	S
	Job string `path:"job"`
}
type ModuleCancelInput struct {
	ModuleJobInput
	Body Empty
}
type Result struct {
	Value any
	ETag  string
}
type Service interface {
	Execute(context.Context, string, any) (Result, error)
	Events(http.ResponseWriter, *http.Request, string)
}
type response[T any] struct {
	ETag string `header:"ETag"`
	Body T
}
type download struct {
	ContentType string `header:"Content-Type"`
	Disposition string `header:"Content-Disposition"`
	Body        []byte
}
type Download struct {
	Data       []byte
	Type, Name string
}
type Route struct {
	Method, Path string
	Queries      []string
}
type API struct {
	Mux    *http.ServeMux
	Huma   huma.API
	Routes []Route
}

func register[I, O any](a *API, service Service, id, method, path string, status int, queries ...string) {
	a.Routes = append(a.Routes, Route{method, path, queries})
	huma.Register(a.Huma, huma.Operation{OperationID: id, Method: method, Path: path, DefaultStatus: status, MaxBodyBytes: 8 << 20, Errors: []int{400, 401, 403, 404, 409, 412, 413, 428, 503}, SkipValidateParams: true}, func(ctx context.Context, input *I) (*response[O], error) {
		if service == nil {
			return nil, Fail(503, "Service unavailable")
		}
		result, err := service.Execute(ctx, id, input)
		if err != nil {
			return nil, err
		}
		body, ok := result.Value.(O)
		if !ok {
			return nil, Fail(500, "Response unavailable")
		}
		return &response[O]{ETag: result.ETag, Body: body}, nil
	})
}
func New(service Service) *API {
	mux := http.NewServeMux()
	config := huma.DefaultConfig("FeCIM local workbench", moduleapi.SchemaVersion)
	config.OpenAPIPath = ""
	config.DocsPath = ""
	config.SchemasPath = ""
	config.CreateHooks = nil
	config.Transformers = []huma.Transformer{func(ctx huma.Context, status string, v any) (any, error) {
		code, _ := strconv.Atoi(status)
		if code >= 400 {
			ctx.SetHeader("Content-Type", "application/problem+json")
			if p, ok := v.(*Problem); ok {
				return p, nil
			}
			return Fail(code, "Request rejected; check typed input and allowed fields"), nil
		}
		return v, nil
	}}
	a := &API{Mux: mux, Huma: humago.New(mux, config)}
	// Worker-only envelopes share the generated Go contract, not HTTP routes.
	a.Huma.OpenAPI().Extensions = map[string]any{"x-worker-envelopes": map[string]any{
		"request": a.Huma.OpenAPI().Components.Schemas.Schema(reflect.TypeFor[moduleapi.ResearchRequest](), true, "ResearchRequest"),
		"reply":   a.Huma.OpenAPI().Components.Schemas.Schema(reflect.TypeFor[moduleapi.ResearchReply](), true, "ResearchReply"),
	}}
	register[BindResourceInput, moduleapi.ResourceReference](a, service, "bind-resource", "POST", "/api/v1/resources", 201)
	register[Nothing, Health](a, service, "health", "GET", "/api/v1/health", 200)
	register[ListProjectsInput, Projects](a, service, "projects", "GET", "/api/v1/projects", 200, "page", "page_size")
	register[CreateProjectInput, Project](a, service, "create-project", "POST", "/api/v1/projects", 201)
	register[OpenProjectInput, Project](a, service, "open-project", "POST", "/api/v1/projects/open", 200)
	register[ProjectInput, Project](a, service, "project", "GET", "/api/v1/projects/{project}", 200)
	register[ProjectInput, Configuration](a, service, "configuration", "GET", "/api/v1/projects/{project}/configuration", 200)
	register[PatchInput, Configuration](a, service, "patch", "PUT", "/api/v1/projects/{project}/configuration", 200)
	register[ProjectMutationInput, Validation](a, service, "validate", "POST", "/api/v1/projects/{project}/validate", 200)
	register[ProjectMutationInput, Status](a, service, "close-project", "POST", "/api/v1/projects/{project}/close", 200)
	register[StartInput, moduleapi.Job](a, service, "start", "POST", "/api/v1/projects/{project}/experiments", 202)
	register[ActiveInput, Jobs](a, service, "active", "GET", "/api/v1/projects/{project}/experiments", 200, "state")
	register[ExecutionInput, moduleapi.Job](a, service, "execution", "GET", "/api/v1/experiments/{execution}", 200)
	register[CancelInput, moduleapi.Job](a, service, "cancel", "POST", "/api/v1/experiments/{execution}/cancel", 200)
	register[RunsInput, study.ResultsPage](a, service, "runs", "GET", "/api/v1/projects/{project}/runs", 200, "page", "page_size", "filter", "sort", "descending", "revision")
	register[RunInput, experiment.RunRecord](a, service, "run", "GET", "/api/v1/projects/{project}/runs/{run}", 200)
	register[AnalysisInput, Analysis](a, service, "analysis", "GET", "/api/v1/projects/{project}/analysis", 200, "page", "page_size", "filter", "sort", "descending", "revision", "compare")
	register[ProjectMutationInput, Analysis](a, service, "verify", "POST", "/api/v1/projects/{project}/verify", 200)
	register[ProjectMutationInput, Artifacts](a, service, "reports", "POST", "/api/v1/projects/{project}/reports", 200)
	register[ExportInput, Artifacts](a, service, "export", "POST", "/api/v1/projects/{project}/exports", 200)
	register[Nothing, Modules](a, service, "modules", "GET", "/api/v1/modules", 200)
	register[ModuleCreateInput, moduleapi.Snapshot](a, service, "module-create", "POST", "/api/v1/modules/{module}/sessions", 201)
	register[ModuleSessionInput, moduleapi.Snapshot](a, service, "module-snapshot", "GET", "/api/v1/module-sessions/{session}", 200)
	register[ModuleActionInput, moduleapi.Snapshot](a, service, "module-action", "POST", "/api/v1/module-sessions/{session}/actions", 200)
	register[ModuleCloseInput, Status](a, service, "module-close", "POST", "/api/v1/module-sessions/{session}/close", 200)
	register[ModuleActionInput, moduleapi.Job](a, service, "module-start", "POST", "/api/v1/module-sessions/{session}/jobs", 202)
	register[ModuleJobInput, moduleapi.Job](a, service, "module-job", "GET", "/api/v1/module-sessions/{session}/jobs/{job}", 200)
	register[ModuleCancelInput, moduleapi.Job](a, service, "module-cancel", "POST", "/api/v1/module-sessions/{session}/jobs/{job}/cancel", 200)
	registerDownload[ReportInput](a, service, "report-download", "/api/v1/projects/{project}/reports/{file}")
	registerDownload[DownloadExportInput](a, service, "export-download", "/api/v1/projects/{project}/exports/{run}/{artifact}")
	// Counts-only SSE is a deliberately separate admission class. Registration
	// remains authoritative here, but stdlib owns flush and write deadlines.
	route := "/api/v1/experiments/{execution}/events"
	a.Routes = append(a.Routes, Route{"GET", route, nil})
	a.Huma.OpenAPI().AddOperation(&huma.Operation{OperationID: "events", Method: "GET", Path: route, Responses: map[string]*huma.Response{"200": {Description: "Counts-only progress; reconnect begins with current snapshot", Content: map[string]*huma.MediaType{"text/event-stream": {Schema: &huma.Schema{Type: "string"}}}}}})
	mux.HandleFunc("GET "+route, func(w http.ResponseWriter, r *http.Request) {
		if service == nil {
			http.Error(w, "unavailable", 503)
			return
		}
		service.Events(w, r, r.PathValue("execution"))
	})
	// Huma's default errors include validation values; the per-API transformer
	// removes them at runtime. Describe that actual safe RFC9457 shape, including
	// typed literal-text conflicts, rather than the framework ErrorModel.
	problemSchema := a.Huma.OpenAPI().Components.Schemas.Schema(reflect.TypeFor[Problem](), true, "Problem")
	for _, item := range a.Huma.OpenAPI().Paths {
		for _, op := range []*huma.Operation{item.Get, item.Post, item.Put} {
			if op == nil {
				continue
			}
			for _, code := range []int{400, 401, 403, 404, 405, 409, 410, 412, 413, 422, 428, 503} {
				op.Responses[strconv.Itoa(code)] = &huma.Response{Description: http.StatusText(code), Content: map[string]*huma.MediaType{"application/problem+json": {Schema: problemSchema}}}
			}
		}
	}
	return a
}
func registerDownload[I any](a *API, s Service, id, path string) {
	a.Routes = append(a.Routes, Route{"GET", path, nil})
	huma.Register(a.Huma, huma.Operation{OperationID: id, Method: "GET", Path: path}, func(ctx context.Context, input *I) (*download, error) {
		r, e := s.Execute(ctx, id, input)
		if e != nil {
			return nil, e
		}
		d := r.Value.(Download)
		return &download{ContentType: d.Type, Disposition: `attachment; filename="` + d.Name + `"`, Body: d.Data}, nil
	})
}

// Policy shares the Go route registry, rejects repeated/unknown queries and
// malformed pagination before handlers. No token or arbitrary URL is allowed.
func (a *API) Policy(method, path string, q url.Values) (known, allowed, queryOK bool) {
	queryOK = true
	for _, r := range a.Routes {
		if !match(r.Path, path) {
			continue
		}
		known = true
		if r.Method != method {
			continue
		}
		allowed = true
		for key, values := range q {
			found := false
			for _, k := range r.Queries {
				found = found || k == key
			}
			if !found || len(values) != 1 {
				queryOK = false
			}
			if key == "page" || key == "page_size" {
				n, e := strconv.Atoi(values[0])
				if e != nil || n < 1 || key == "page_size" && n > 100 {
					queryOK = false
				}
			}
			if key == "descending" && values[0] != "true" && values[0] != "false" {
				queryOK = false
			}
			if key == "state" && values[0] != "active" {
				queryOK = false
			}
		}
		return
	}
	if len(q) > 0 {
		queryOK = false
	}
	return
}
func match(pattern, path string) bool {
	a, b := strings.Split(pattern, "/"), strings.Split(path, "/")
	if len(a) != len(b) {
		return false
	}
	for i, p := range a {
		if strings.HasPrefix(p, "{") {
			if b[i] == "" {
				return false
			}
			continue
		}
		if p != b[i] {
			return false
		}
	}
	return true
}
func (a *API) JSON() ([]byte, error) { return json.MarshalIndent(a.Huma.OpenAPI(), "", "  ") }
