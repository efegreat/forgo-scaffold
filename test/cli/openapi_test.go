package cli

import (
	"os"
	"path/filepath"
	"testing"

	"example/test/support"
)

// golden is the API document of version v1 the repository holds, which clients read.
const golden = "../../api/openapi-v1.json"

// The openapi command documents config.HTTP as api/openapi-v1.json holds it. When the API changes, the
// document changes with it, in the same pull request:
//
//	go run ./cmd/cli openapi -o=api/openapi.json
func TestOpenAPI(t *testing.T) {
	out := filepath.Join(t.TempDir(), "openapi.json")
	if err := support.Command("openapi", "-o="+out); err != nil {
		t.Fatal(err)
	}

	got, err := os.ReadFile(filepath.Join(filepath.Dir(out), "openapi-v1.json"))
	if err != nil {
		t.Fatal(err)
	}
	want, err := os.ReadFile(golden)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != string(want) {
		t.Fatalf("%s is out of date: run go run ./cmd/cli openapi -o=api/openapi.json from the root of the module", golden)
	}
}
