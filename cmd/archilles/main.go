// Command archilles checks a repository against its own architecture.
//
//	archilles check                 the standing check: the map matches the territory
//	archilles plan --base <ref>     what a change does, what it touches, what it breaks
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"archilles/pkg/adapter/git"
	"archilles/pkg/adapter/golang"
	"archilles/pkg/design"
)

const liveRel = "archilles/architecture-live"

func main() {
	root := "."
	cmd := "check"
	args := os.Args[1:]
	if len(args) > 0 && !strings.HasPrefix(args[0], "-") {
		cmd, args = args[0], args[1:]
	}
	fs := flag.NewFlagSet(cmd, flag.ExitOnError)
	base := fs.String("base", "", "ref to plan against (plan only)")
	fs.Parse(args)

	in := design.Input{Today: time.Now().UTC(), LiveRel: liveRel}
	var err error
	if in.Records, err = design.Load(root, liveRel); err != nil {
		fail(err)
	}
	if in.Packages, err = packages(root); err != nil {
		fail(err)
	}

	switch cmd {
	case "check":
	case "plan":
		if *base == "" {
			fail(fmt.Errorf("plan needs --base <ref>"))
		}
		if in.Changes, in.Commits, err = history(root, *base); err != nil {
			fail(err)
		}
	default:
		fail(fmt.Errorf("unknown command %q", cmd))
	}

	findings := design.Plan(in)
	report(in, findings)
	if design.Open(findings) > 0 {
		os.Exit(1)
	}
}

// packages runs the Go adapter and hands design plain directories.
func packages(root string) ([]design.Package, error) {
	pkgs, err := golang.Extract(root)
	if err != nil {
		return nil, err
	}
	dirOf := map[string]string{}
	for _, p := range pkgs {
		dirOf[p.ImportPath] = p.Dir
	}
	out := make([]design.Package, 0, len(pkgs))
	for _, p := range pkgs {
		d := design.Package{Dir: p.Dir}
		for _, imp := range p.Imports {
			d.Imports = append(d.Imports, dirOf[imp])
		}
		out = append(out, d)
	}
	return out, nil
}

// history reads what changed since base, with both versions of each record.
func history(root, base string) ([]design.Change, []design.Commit, error) {
	changed, err := git.Changed(root, base, "HEAD")
	if err != nil {
		return nil, nil, err
	}
	var changes []design.Change
	for _, c := range changed {
		d := design.Change{Status: c.Status, Path: c.Path, OldPath: c.OldPath}
		if strings.HasPrefix(c.Path, liveRel+"/") {
			old := c.Path
			if c.OldPath != "" {
				old = c.OldPath
			}
			if d.Old, err = git.Show(root, base, old); err != nil {
				return nil, nil, err
			}
			if c.Status != "D" {
				if d.New, err = os.ReadFile(filepath.Join(root, c.Path)); err != nil {
					return nil, nil, err
				}
			}
		}
		changes = append(changes, d)
	}
	commits, err := git.Commits(root, base, "HEAD")
	if err != nil {
		return nil, nil, err
	}
	var out []design.Commit
	for _, c := range commits {
		out = append(out, design.Commit(c))
	}
	return changes, out, nil
}

func report(in design.Input, findings []design.Finding) {
	comps := 0
	for _, r := range in.Records {
		if r.Kind == "component" {
			comps++
		}
	}
	fmt.Printf("components  %d\npackages    %d\n", comps, len(in.Packages))

	if len(in.Changes) > 0 {
		fmt.Println("\nrecords changed")
		any := false
		for _, c := range in.Changes {
			if !strings.HasPrefix(c.Path, liveRel+"/") {
				continue
			}
			any = true
			p := strings.TrimPrefix(c.Path, liveRel+"/")
			if c.OldPath != "" {
				p = strings.TrimPrefix(c.OldPath, liveRel+"/") + " -> " + p
			}
			fmt.Printf("  %s  %s\n", c.Status, p)
		}
		if !any {
			fmt.Println("  none")
		}
	}

	fmt.Println()
	if len(findings) == 0 {
		fmt.Println("findings    none")
	}
	last := ""
	for _, f := range findings {
		if f.Kind != last {
			fmt.Println(f.Kind)
			last = f.Kind
		}
		mark := "open"
		if f.State == "accepted" {
			mark = "accepted"
		}
		line := "  " + strings.TrimPrefix(f.Subject, liveRel+"/")
		if f.Detail != "" {
			line += "   " + f.Detail
		}
		fmt.Printf("%-64s %s\n", line, mark)
	}
	fmt.Printf("\nopen        %d\n", design.Open(findings))
}

func fail(err error) {
	fmt.Fprintln(os.Stderr, "archilles:", err)
	os.Exit(2)
}
