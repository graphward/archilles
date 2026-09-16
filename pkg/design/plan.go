package design

import (
	"sort"
	"strings"
	"time"

	"github.com/goccy/go-yaml"
)

// Package is a unit of code the adapter found: where it is and what it
// imports, both as directories relative to the repo root.
type Package struct {
	Dir     string
	Imports []string
}

// Change is one path that differs between the base and head of a change,
// with both versions of the content where they exist.
type Change struct {
	Status  string // A M D R
	Path    string
	OldPath string
	Old     []byte
	New     []byte
}

// Commit is what a commit said. The body is the why.
type Commit struct {
	Hash    string
	Subject string
	Body    string
}

// Input is everything a plan is computed from. Nothing here is fetched; the
// caller read it.
type Input struct {
	Records  []Record
	Packages []Package
	Changes  []Change // empty for a plain check
	Commits  []Commit // empty for a plain check
	Today    time.Time
	LiveRel  string
}

// Finding is one thing to look at. State is open, or accepted when an
// unexpired exception covers it - the same shape as a security finding with
// an accepted-risk justification.
type Finding struct {
	Kind    string // unassigned missing dangling modified dependent no-why expired
	Subject string
	Detail  string
	State   string // open | accepted
}

// Plan computes every finding over the input. With no Changes or Commits it
// is the standing check; with them it is the plan for a change.
func Plan(in Input) []Finding {
	var out []Finding
	refs := Refs(in.Records)
	exceptions := activeExceptions(in.Records, in.Today)

	// --- drift: the map matches the territory ------------------------------
	comps := filter(in.Records, "component")
	owner := map[string]string{}
	used := map[string]bool{}
	for _, p := range in.Packages {
		best := -1
		for _, c := range comps {
			cp := c.Fields["path"]
			if (p.Dir == cp || strings.HasPrefix(p.Dir, cp+"/")) && len(cp) > best {
				best = len(cp)
				owner[p.Dir] = c.Name
			}
		}
		if best < 0 {
			f := Finding{Kind: "unassigned", Subject: p.Dir, State: "open"}
			if ex := exceptions["package-coverage"][p.Dir]; ex != nil {
				f.State, f.Detail = "accepted", "exception "+ex.Name+" until "+ex.Fields["expires"]
			}
			out = append(out, f)
		} else {
			used[owner[p.Dir]] = true
		}
	}
	for _, c := range comps {
		if !used[c.Name] {
			f := Finding{Kind: "missing", Subject: c.Name, Detail: c.Fields["path"], State: "open"}
			if ex := exceptions["package-coverage"][c.Fields["path"]]; ex != nil {
				f.State, f.Detail = "accepted", "exception "+ex.Name+" until "+ex.Fields["expires"]
			}
			out = append(out, f)
		}
	}

	// --- references resolve --------------------------------------------------
	for _, r := range refs {
		if r.Res == nil {
			out = append(out, Finding{Kind: "dangling", Subject: r.From.Path,
				Detail: r.Field + ": " + r.To, State: "open"})
		}
	}

	// --- exceptions expire ---------------------------------------------------
	for _, e := range filter(in.Records, "exception") {
		if exp, ok := parseDate(e.Fields["expires"]); ok && !exp.After(in.Today) {
			out = append(out, Finding{Kind: "expired", Subject: e.Path, Detail: "expired " + e.Fields["expires"], State: "open"})
		}
	}

	if len(in.Changes) == 0 && len(in.Commits) == 0 {
		return sorted(out)
	}

	// --- records are immutable; only anchors move ----------------------------
	inPR := map[string]bool{}
	var changedNames []string
	for _, c := range in.Changes {
		if !strings.HasPrefix(c.Path, in.LiveRel+"/") {
			continue
		}
		inPR[c.Path] = true
		if c.OldPath != "" {
			inPR[c.OldPath] = true
		}
		rel := strings.TrimPrefix(c.Path, in.LiveRel+"/")
		kind, name := Classify(rel)
		if kind == "" {
			continue
		}
		changedNames = append(changedNames, name)
		if c.Status == "M" && !anchorOnly(c.Old, c.New) {
			out = append(out, Finding{Kind: "modified", Subject: c.Path,
				Detail: "edited in place; add a record that supersedes it", State: "open"})
		}
	}

	// --- a changed record drags its dependents into the PR -------------------
	for _, name := range changedNames {
		for _, dep := range Dependents(refs, name) {
			if !inPR[dep.Path] {
				out = append(out, Finding{Kind: "dependent", Subject: dep.Path,
					Detail: "references " + name + ", which changed; not in this change", State: "open"})
			}
		}
	}

	// --- every commit carries a why ------------------------------------------
	for _, c := range in.Commits {
		if strings.TrimSpace(c.Body) == "" {
			out = append(out, Finding{Kind: "no-why", Subject: c.Hash[:7], Detail: c.Subject, State: "open"})
		}
	}

	return sorted(out)
}

// Open counts findings that are not accepted.
func Open(fs []Finding) int {
	n := 0
	for _, f := range fs {
		if f.State == "open" {
			n++
		}
	}
	return n
}

func filter(recs []Record, kind string) []Record {
	var out []Record
	for _, r := range recs {
		if r.Kind == kind {
			out = append(out, r)
		}
	}
	return out
}

// activeExceptions indexes unexpired exceptions by rule, then by path.
func activeExceptions(recs []Record, today time.Time) map[string]map[string]*Record {
	idx := map[string]map[string]*Record{}
	for i := range recs {
		e := recs[i]
		if e.Kind != "exception" {
			continue
		}
		if exp, ok := parseDate(e.Fields["expires"]); ok && !exp.After(today) {
			continue
		}
		rule := e.Fields["rule"]
		if idx[rule] == nil {
			idx[rule] = map[string]*Record{}
		}
		idx[rule][e.Fields["path"]] = &recs[i]
	}
	return idx
}

// anchorOnly is true when the only field that differs is path.
func anchorOnly(old, new []byte) bool {
	var a, b map[string]any
	if yaml.Unmarshal(old, &a) != nil || yaml.Unmarshal(new, &b) != nil {
		return false
	}
	if len(a) != len(b) {
		return false
	}
	for k, av := range a {
		bv, ok := b[k]
		if !ok {
			return false
		}
		if k == "path" {
			continue
		}
		if fmtv(av) != fmtv(bv) {
			return false
		}
	}
	return true
}

func fmtv(v any) string {
	b, _ := yaml.Marshal(v)
	return string(b)
}

func parseDate(s string) (time.Time, bool) {
	t, err := time.Parse("2006-01-02", s)
	return t, err == nil
}

func sorted(fs []Finding) []Finding {
	sort.SliceStable(fs, func(i, j int) bool {
		if fs[i].Kind != fs[j].Kind {
			return fs[i].Kind < fs[j].Kind
		}
		return fs[i].Subject < fs[j].Subject
	})
	return fs
}
