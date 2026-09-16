// Package design is the model of intent and everything computed over it. It
// reads the live tree, resolves references between records by name, and
// produces findings. It knows nothing about git, Go, or any toolchain: the
// caller hands it what those adapters read.
package design

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/goccy/go-yaml"
)

// Record is one file in the live tree. Kind comes from where it sits, name
// from the directory (components) or the filename (everything else).
type Record struct {
	Kind   string            // org | system | component | rule | exception | decision
	Name   string            //
	Scope  string            // directory containing it, relative to the live root
	Path   string            // file path relative to the repo root
	Fields map[string]string // flat scalar fields
}

// refFields are the only fields that name another record. Anything mentioned
// only in prose is, by definition, not a relationship.
var refFields = map[string]string{
	"rule":       "rule", // must resolve to a rule
	"supersedes": "",     // resolves to the same kind as the referrer
	"cites":      "*",    // any kind
}

// Classify says what a live-tree path is, or "" if it is not a record.
func Classify(rel string) (kind, name string) {
	dir, file := filepath.Split(filepath.ToSlash(rel))
	dir = strings.TrimSuffix(dir, "/")
	stem := strings.TrimSuffix(strings.TrimSuffix(file, ".yaml"), ".md")
	switch {
	case file == "component.yaml":
		return "component", filepath.Base(dir)
	case file == "system.yaml":
		return "system", filepath.Base(dir)
	case file == "org.yaml":
		return "org", filepath.Base(dir)
	case filepath.Base(dir) == "rules" && strings.HasSuffix(file, ".yaml"):
		return "rule", stem
	case filepath.Base(dir) == "exceptions" && strings.HasSuffix(file, ".yaml"):
		return "exception", stem
	case filepath.Base(dir) == "decisions" && strings.HasSuffix(file, ".md"):
		return "decision", stem
	}
	return "", ""
}

// Load reads every record under the live tree.
func Load(root, liveRel string) ([]Record, error) {
	live := filepath.Join(root, liveRel)
	var recs []Record
	err := filepath.WalkDir(live, func(p string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		rel, _ := filepath.Rel(live, p)
		kind, name := Classify(rel)
		if kind == "" {
			return nil
		}
		b, err := os.ReadFile(p)
		if err != nil {
			return err
		}
		fields, err := scalars(b)
		if err != nil {
			return fmt.Errorf("%s: %w", rel, err)
		}
		recs = append(recs, Record{
			Kind:   kind,
			Name:   name,
			Scope:  filepath.ToSlash(filepath.Dir(scopeOf(rel, kind))),
			Path:   filepath.ToSlash(filepath.Join(liveRel, rel)),
			Fields: fields,
		})
		return nil
	})
	sort.Slice(recs, func(i, j int) bool { return recs[i].Path < recs[j].Path })
	return recs, err
}

// scopeOf strips the typed directory so a rule in x/rules/ is scoped to x/.
func scopeOf(rel, kind string) string {
	switch kind {
	case "rule", "exception", "decision":
		return filepath.Dir(rel)
	}
	return rel
}

// scalars pulls the top-level string fields out of a YAML document. Markdown
// records have none.
func scalars(b []byte) (map[string]string, error) {
	out := map[string]string{}
	if !strings.HasPrefix(strings.TrimSpace(string(b)), "#") && strings.Contains(string(b), ":") {
		var m map[string]any
		if err := yaml.Unmarshal(b, &m); err != nil {
			return nil, err
		}
		for k, v := range m {
			switch x := v.(type) {
			case string:
				out[k] = x
			case fmt.Stringer:
				out[k] = x.String()
			case time.Time:
				out[k] = x.Format("2006-01-02")
			}
		}
	}
	return out, nil
}

// Ref is one record naming another.
type Ref struct {
	From  Record
	Field string
	To    string  // the name as written
	Res   *Record // nil when dangling
}

// Refs resolves every reference field across the tree, by name.
func Refs(recs []Record) []Ref {
	var refs []Ref
	for _, r := range recs {
		for field, wantKind := range refFields {
			to, ok := r.Fields[field]
			if !ok {
				continue
			}
			kind := wantKind
			if kind == "" {
				kind = r.Kind
			}
			ref := Ref{From: r, Field: field, To: to}
			for i := range recs {
				if recs[i].Name == to && (kind == "*" || recs[i].Kind == kind) {
					ref.Res = &recs[i]
					break
				}
			}
			refs = append(refs, ref)
		}
	}
	return refs
}

// Dependents returns every record that references the named one.
func Dependents(refs []Ref, name string) []Record {
	var out []Record
	for _, ref := range refs {
		if ref.Res != nil && ref.Res.Name == name {
			out = append(out, ref.From)
		}
	}
	return out
}
