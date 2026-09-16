// Package git reads history from a git repository. It is the read side for
// everything archilles derives from version control: what changed between
// two refs, what a file looked like at a ref, and what the commits said.
package git

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

// Change is one path's status between two refs, as `git diff --name-status -M`
// reports it: A added, M modified, D deleted, R renamed (OldPath set).
type Change struct {
	Status  string
	Path    string
	OldPath string
}

// Commit is one commit's message: subject, body, and the trailers at the
// end of the body as git interpret-trailers would read them.
type Commit struct {
	Hash     string
	Subject  string
	Body     string
	Trailers map[string][]string
}

func run(root string, args ...string) ([]byte, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = root
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("git %s: %w\n%s", strings.Join(args, " "), err, stderr.String())
	}
	return out, nil
}

// Changed lists paths that differ between base and head, rename-aware.
func Changed(root, base, head string) ([]Change, error) {
	out, err := run(root, "diff", "--name-status", "-M", base+"..."+head)
	if err != nil {
		return nil, err
	}
	var changes []Change
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		if line == "" {
			continue
		}
		f := strings.Split(line, "\t")
		c := Change{Status: f[0][:1], Path: f[1]}
		if c.Status == "R" && len(f) > 2 {
			c.OldPath, c.Path = f[1], f[2]
		}
		changes = append(changes, c)
	}
	return changes, nil
}

// Show returns a file's content at a ref. A missing file is nil, not an error.
func Show(root, ref, path string) ([]byte, error) {
	out, err := run(root, "show", ref+":"+path)
	if err != nil {
		if strings.Contains(err.Error(), "does not exist") || strings.Contains(err.Error(), "exists on disk, but not in") {
			return nil, nil
		}
		return nil, err
	}
	return out, nil
}

// Commits returns the non-merge commits in base..head, oldest first.
func Commits(root, base, head string) ([]Commit, error) {
	out, err := run(root, "log", "--no-merges", "--reverse", "--format=%H%x1f%s%x1f%b%x1e", base+".."+head)
	if err != nil {
		return nil, err
	}
	var commits []Commit
	for _, rec := range strings.Split(string(out), "\x1e") {
		f := strings.SplitN(strings.TrimSpace(rec), "\x1f", 3)
		if len(f) < 2 || f[0] == "" {
			continue
		}
		c := Commit{Hash: f[0], Subject: f[1], Trailers: map[string][]string{}}
		if len(f) == 3 {
			c.Body, c.Trailers = splitTrailers(strings.TrimSpace(f[2]))
		}
		commits = append(commits, c)
	}
	return commits, nil
}

// splitTrailers separates the trailer block - the final paragraph, when every
// line of it is Key: value - from the rest of the body.
func splitTrailers(body string) (string, map[string][]string) {
	trailers := map[string][]string{}
	paras := strings.Split(body, "\n\n")
	if len(paras) == 0 {
		return body, trailers
	}
	last := strings.Split(strings.TrimSpace(paras[len(paras)-1]), "\n")
	for _, l := range last {
		k, v, ok := strings.Cut(l, ":")
		if !ok || strings.ContainsAny(k, " \t") || k == "" {
			return body, map[string][]string{}
		}
		trailers[k] = append(trailers[k], strings.TrimSpace(v))
	}
	return strings.TrimSpace(strings.Join(paras[:len(paras)-1], "\n\n")), trailers
}
