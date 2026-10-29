package project

import (
	"os"
	"reflect"
	"strings"
	"testing"
)

// Exposing an unallowlisted field, alias or caller-mutable snapshot, or parsing
// formatted numeric strings in the client, would break this editor boundary.
func TestConfigurationFields(t *testing.T) {
	docs := map[string][]byte{"project.yaml": []byte(validProjectYAML), "design.yaml": []byte(validDesignYAML), "sweep.yaml": []byte(validSweepYAML)}
	docs["design.yaml"] = []byte(strings.Replace(validDesignYAML, "rows: 32", "rows: 0x20", 1))
	snapshot, err := ReadConfigurationBytes(docs)
	if err != nil {
		t.Fatal(err)
	}
	fields, err := snapshot.Fields()
	if err != nil {
		t.Fatal(err)
	}
	got := map[string]ConfigurationField{}
	for _, field := range fields {
		got[field.Document+field.Path] = field
	}
	for _, tt := range []ConfigurationField{
		{Document: "project.yaml", Path: "/name", Kind: "string", Value: "HZO study"},
		{Document: "design.yaml", Path: "/array/rows", Kind: "int", Value: "32"},
		{Document: "design.yaml", Path: "/device/g_min_s", Kind: "float", Value: "1e-06"},
		{Document: "sweep.yaml", Path: "/parameters/1/range/count", Kind: "int", Value: "2"},
		{Document: "sweep.yaml", Path: "/parameters/0/values/0", Kind: "float", Value: "16"},
	} {
		if field, ok := got[tt.Document+tt.Path]; !ok || field != tt {
			t.Fatalf("want %+v got %+v", tt, field)
		}
	}
	for _, key := range []string{"project.yaml/id", "project.yaml/model_version", "project.yaml/citations/0", "project.yaml/objectives/0/metric", "sweep.yaml/parameters/0/path"} {
		if _, ok := got[key]; ok {
			t.Fatalf("immutable field exposed: %s", key)
		}
	}
	dir := t.TempDir()
	for name, data := range docs {
		if err := os.WriteFile(dir+"/"+name, data, 0600); err != nil {
			t.Fatal(err)
		}
	}
	root, err := os.OpenRoot(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer root.Close()
	native, err := ReadConfiguration(root)
	if err != nil {
		t.Fatal(err)
	}
	nativeFields, err := native.Fields()
	if err != nil || native.Revision() != snapshot.Revision() || !reflect.DeepEqual(fields, nativeFields) {
		t.Fatalf("native/bytes fields mismatch: %v", err)
	}
	docs["design.yaml"][0] = 'X'
	fields[0].Value = "changed"
	again, err := snapshot.Fields()
	if err != nil || reflect.DeepEqual(fields, again) {
		t.Fatalf("snapshot was caller-mutable: %v", err)
	}
	for _, body := range []string{validDesignYAML + "---\n", validDesignYAML + "extra: &x 1\ncopy: *x\n", strings.Repeat("#", 1<<20) + "\n"} {
		bad := map[string][]byte{"project.yaml": []byte(validProjectYAML), "design.yaml": []byte(body), "sweep.yaml": []byte(validSweepYAML)}
		candidate, readErr := ReadConfigurationBytes(bad)
		if readErr == nil {
			if _, err := candidate.Fields(); err == nil {
				t.Fatal("ambiguous or oversize snapshot exposed")
			}
		}
	}
	if _, err := (ConfigurationSnapshot{}).Fields(); err == nil {
		t.Fatal("zero snapshot exposed")
	}
	if _, err := ReadConfigurationBytes(map[string][]byte{}); err == nil {
		t.Fatal("missing documents accepted")
	}
}
