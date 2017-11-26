package gig

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
)

type Ref interface {
	Repo() *Repository
	Name() string
	Fullname() string
	Namespace() string
	Resolve() (SHA1, error)
}

type ref struct {
	repo *Repository
	name string
	ns   string // #special, #branch, or the name like 'remote', 'tags'
}

func (r *ref) Name() string {
	return r.name
}

func (r *ref) Fullname() string {
	fullname := r.name
	if !strings.HasPrefix(r.ns, "#") {
		fullname = path.Join(r.ns, r.name)
	}
	return fullname
}

func (r *ref) Repo() *Repository {
	return r.repo
}

func (r *ref) Namespace() string {
	return r.ns
}

func IsBranchRef(r Ref) bool {
	return r.Namespace() == "#branch"
}

//IDRef is a reference that points via
//a sha1 directly to a git object
type IDRef struct {
	ref
	id SHA1
}

//Resolve for IDRef returns the stored object
//id (SHA1)
func (r *IDRef) Resolve() (SHA1, error) {
	return r.id, nil
}

//SymbolicRef is a reference that points
//to another reference
type SymbolicRef struct {
	ref
	Symbol string
}

//Resolve will resolve the symbolic reference into
//an object id.
func (r *SymbolicRef) Resolve() (SHA1, error) {
	gdir := fmt.Sprintf("--git-dir=%s", r.repo.Path)

	cmd := exec.Command("git", gdir, "rev-parse", r.Fullname())
	body, err := cmd.Output()

	if err != nil {
		var id SHA1
		return id, err
	}

	return ParseSHA1(string(body))
}

func parseRefName(filename string) (name, ns string, err error) {
	comps := strings.Split(filename, "/")
	n := len(comps)

	if n < 1 || n == 2 || (n > 2 && comps[0] != "refs") {
		err = fmt.Errorf("git: unexpected ref name: %v", filename)
		return
	}

	if n == 1 {
		name = comps[0]
		ns = "#special"
	}

	// 'man gitrepository-layout' is really helpfull
	// 'man git-check-ref-format' too
	// [HEAD|ORIG_HEAD] -> special head
	// [0|refs][1|<ns>][2+|name]
	// <ns> == "heads" -> local branch"
	switch {
	case n == 1:
		name = comps[0]
		ns = "#special"
	case comps[1] == "heads":
		name = path.Join(comps[2:]...)
		ns = "#branch"
	default:
		name = path.Join(comps[2:]...)
		ns = comps[1]
	}
	return
}

func (repo *Repository) parseRef(filename string) (Ref, error) {

	name, ns, err := parseRefName(filename)
	if err != nil {
		return nil, err
	}

	base := ref{repo, name, ns}

	//now to the actual contents of the ref
	data, err := ioutil.ReadFile(filepath.Join(repo.Path, filename))
	if err != nil {
		if os.IsNotExist(err) {
			return repo.findPackedRef(base.Fullname())
		}
		return nil, err
	}

	b := string(data)
	if strings.HasPrefix(b, "ref:") {
		trimmed := strings.Trim(b[4:], " \n")
		return &SymbolicRef{base, trimmed}, nil
	}

	id, err := ParseSHA1(b)
	if err == nil {
		return &IDRef{base, id}, nil
	}

	return nil, fmt.Errorf("git: unknown ref type: %q", b)
}

func (repo *Repository) listRefWithName(name string) (res []Ref) {
	gdir := fmt.Sprintf("--git-dir=%s", repo.Path)
	cmd := exec.Command("git", gdir, "show-ref", name)
	body, err := cmd.Output()

	if err != nil {
		return
	}

	r := bytes.NewBuffer(body)

	for {
		var l string
		l, err = r.ReadString('\n')
		if err != nil {
			break
		}

		_, name := split2(l[:len(l)-1], " ")
		r, err := repo.parseRef(name)

		if err != nil {
			fmt.Fprintf(os.Stderr, "git: could not parse ref with name %q: %v", name, err)
			continue
		}

		res = append(res, r)
	}

	return
}

func (repo *Repository) loadPackedRefs() ([]Ref, error) {

	fd, err := os.Open(filepath.Join(repo.Path, "packed-refs"))
	if err != nil {
		return nil, err
	}
	defer fd.Close()

	r := bufio.NewReader(fd)

	var refs []Ref
	for {
		var l string
		l, err = r.ReadString('\n')
		if err != nil {
			break
		}

		head, tail := split2(l, " ")
		if tail == "" {
			//probably a peeled id (i.e. "^SHA1")
			//TODO: do something with it
			continue
		}

		name, ns, err := parseRefName(tail[:len(tail)-1])
		if err != nil {
			//TODO: log error, panic?
			continue
		}

		id, err := ParseSHA1(head)
		if err != nil {
			//TODO: same as above
			continue
		}

		refs = append(refs, &IDRef{ref{repo, name, ns}, id})
	}

	if err != nil && err != io.EOF {
		return nil, err
	}

	return refs, nil
}

func (repo *Repository) findPackedRef(name string) (Ref, error) {
	refs, err := repo.loadPackedRefs()
	if err != nil {
		return nil, err
	}

	for _, ref := range refs {
		if ref.Fullname() == name {
			return ref, nil
		}
	}
	return nil, fmt.Errorf("ref with name %q not found", name)
}
