package example_test

import (
	"go/parser"
	"go/token"
	"io/fs"
	"path"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"testing"
)

// module is the module path, from go.mod.
const module = "example"

// layers are the directories of internal, and the layers each one may import. Inside a layer, a domain
// or an adapter, such as domain/order, may import its own subpackages but never a sibling: domains
// share entity, errs, event and port instead of importing each other.
var layers = map[string][]string{
	"entity":  {"errs"}, // an entity's invariants fail with business errors
	"errs":    nil,
	"event":   nil,
	"port":    {"entity", "event"},
	"domain":  {"entity", "errs", "event", "port"},
	"adapter": {"entity", "errs", "event", "port"},
	"app":     {"adapter", "domain", "entity", "errs", "event", "port"},
}

// infrastructure are the packages that only adapter and app may import: the rest of the application
// never knows which database or service is behind a port. Add the clients your adapters use.
var infrastructure = []string{"database/sql", "github.com/efegreat/forgo-adapter"}

// unit returns the part of internal that pkg, a directory of internal, belongs to: a domain or an
// adapter, such as domain/order, or else its layer.
func unit(pkg string) string {
	parts := strings.SplitN(pkg, "/", 3)
	if (parts[0] == "domain" || parts[0] == "adapter") && len(parts) > 1 {
		return parts[0] + "/" + parts[1]
	}
	return parts[0]
}

// The packages of internal import only what their layer may: the dependencies point from app, through
// adapter and domain, to port and the shared packages, and never back. The compiler cannot see this
// until it becomes an import cycle, so the test fails on the first import across the rule instead.
func TestLayers(t *testing.T) {
	files := 0
	err := filepath.WalkDir("internal", func(file string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || !strings.HasSuffix(file, ".go") || strings.HasSuffix(file, "_test.go") {
			return err
		}
		files++
		pkg, _ := strings.CutPrefix(path.Dir(filepath.ToSlash(file)), "internal")
		pkg = strings.TrimPrefix(pkg, "/")
		layer, _, _ := strings.Cut(pkg, "/")
		allowed, ok := layers[layer]
		if !ok {
			t.Errorf("%s: %s is no layer: move it to one, or add it to layers", file, path.Join("internal", layer))
			return nil
		}

		f, err := parser.ParseFile(token.NewFileSet(), file, nil, parser.ImportsOnly)
		if err != nil {
			return err
		}
		for _, spec := range f.Imports {
			imp, err := strconv.Unquote(spec.Path.Value)
			if err != nil {
				return err
			}
			if target, ok := strings.CutPrefix(imp, module+"/internal/"); ok {
				targetLayer, _, _ := strings.Cut(target, "/")
				if unit(target) != unit(pkg) && !slices.Contains(allowed, targetLayer) {
					t.Errorf("%s imports %s: %s may import only %v", file, imp, layer, allowed)
				}
				continue
			}
			if layer == "adapter" || layer == "app" {
				continue
			}
			for _, infra := range infrastructure {
				if imp == infra || strings.HasPrefix(imp, infra+"/") {
					t.Errorf("%s imports %s: only adapter and app may", file, imp)
				}
			}
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if files == 0 {
		t.Fatal("no files under internal: run the test from the root of the module")
	}
}
