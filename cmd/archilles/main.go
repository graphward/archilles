// Command archilles checks a Go module against its own architecture.
package main

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/goccy/go-yaml"

	"archilles/pkg/adapter/golang"
)

type component struct {
	Metadata struct {
		Name string `yaml:"name"`
	} `yaml:"metadata"`
	Spec struct {
		Layer string `yaml:"layer"`
		Repo  string `yaml:"repo"`
		Path  string `yaml:"path"`
	} `yaml:"spec"`
}

func main() {
	root := "."
	comps, err := loadComponents(filepath.Join(root, "archilles", "architecture-live"))
	if err != nil {
		fail(err)
	}
	pkgs, err := golang.Extract(root)
	if err != nil {
		fail(err)
	}

	// A package belongs to the component whose path is its longest prefix.
	owner := map[string]string{}
	used := map[string]bool{}
	var unassigned []string
	for _, p := range pkgs {
		best := -1
		for _, c := range comps {
			if (p.Dir == c.Spec.Path || strings.HasPrefix(p.Dir, c.Spec.Path+"/")) && len(c.Spec.Path) > best {
				best = len(c.Spec.Path)
				owner[p.Dir] = c.Metadata.Name
			}
		}
		if best < 0 {
			unassigned = append(unassigned, p.Dir)
		} else {
			used[owner[p.Dir]] = true
		}
	}

	var missing []string
	for _, c := range comps {
		if !used[c.Metadata.Name] {
			missing = append(missing, c.Metadata.Name)
		}
	}

	// Package imports collapse to component edges.
	dirOf := map[string]string{}
	for _, p := range pkgs {
		dirOf[p.ImportPath] = p.Dir
	}
	edgeSet := map[string]bool{}
	for _, p := range pkgs {
		from := owner[p.Dir]
		for _, imp := range p.Imports {
			to := owner[dirOf[imp]]
			if from != "" && to != "" && from != to {
				edgeSet[from+" -> "+to] = true
			}
		}
	}
	var edges []string
	for e := range edgeSet {
		edges = append(edges, e)
	}
	sort.Strings(edges)

	fmt.Printf("components  %d\n", len(comps))
	fmt.Printf("packages    %d\n", len(pkgs))
	fmt.Printf("unassigned  %d\n", len(unassigned))
	for _, u := range unassigned {
		fmt.Printf("  %s\n", u)
	}
	fmt.Printf("missing     %d\n", len(missing))
	for _, m := range missing {
		fmt.Printf("  %s\n", m)
	}
	fmt.Println()
	for _, e := range edges {
		fmt.Println(e)
	}

	if len(unassigned)+len(missing) > 0 {
		os.Exit(1)
	}
}

// loadComponents reads every component.yaml under the live tree.
func loadComponents(live string) ([]component, error) {
	var comps []component
	err := filepath.WalkDir(live, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || d.Name() != "component.yaml" {
			return err
		}
		b, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		var c component
		if err := yaml.Unmarshal(b, &c); err != nil {
			return fmt.Errorf("%s: %w", path, err)
		}
		comps = append(comps, c)
		return nil
	})
	sort.Slice(comps, func(i, j int) bool { return comps[i].Metadata.Name < comps[j].Metadata.Name })
	return comps, err
}

func fail(err error) {
	fmt.Fprintln(os.Stderr, "archilles:", err)
	os.Exit(2)
}
