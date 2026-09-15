// Package golang extracts the package graph of a Go module.
package golang

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os/exec"
	"path/filepath"
	"strings"
)

// Package is one Go package in the module and the module-internal packages it imports.
type Package struct {
	ImportPath string
	Dir        string   // relative to the module root, slash-separated
	Imports    []string // import paths within the same module
}

type listed struct {
	ImportPath string
	Dir        string
	Imports    []string
	Module     *struct{ Path, Dir string }
}

// Extract runs `go list -json ./...` at root and keeps only module-internal imports.
func Extract(root string) ([]Package, error) {
	cmd := exec.Command("go", "list", "-json", "./...")
	cmd.Dir = root
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("go list: %w\n%s", err, stderr.String())
	}

	dec := json.NewDecoder(bytes.NewReader(out))
	var pkgs []Package
	for {
		var l listed
		if err := dec.Decode(&l); err == io.EOF {
			break
		} else if err != nil {
			return nil, err
		}
		if l.Module == nil {
			continue
		}
		rel, err := filepath.Rel(l.Module.Dir, l.Dir)
		if err != nil {
			return nil, err
		}
		p := Package{ImportPath: l.ImportPath, Dir: filepath.ToSlash(rel)}
		for _, imp := range l.Imports {
			if strings.HasPrefix(imp, l.Module.Path+"/") {
				p.Imports = append(p.Imports, imp)
			}
		}
		pkgs = append(pkgs, p)
	}
	return pkgs, nil
}
