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

//Repository represents an on disk git repository.
type Repository struct {
	Path string
}

//InitBareRepository creates a bare git repository at path.
func InitBareRepository(path string) (*Repository, error) {

	path, err := filepath.Abs(path)
	if err != nil {
		return nil, fmt.Errorf("Could not determine absolute path: %v", err)
	}

	cmd := exec.Command("git", "init", "--bare", path)
	err = cmd.Run()

	if err != nil {
		return nil, err
	}

	return &Repository{Path: path}, nil
}

//IsBareRepository checks if path is a bare git repository.
func IsBareRepository(path string) bool {

	cmd := exec.Command("git", fmt.Sprintf("--git-dir=%s", path), "rev-parse", "--is-bare-repository")
	body, err := cmd.Output()

	if err != nil {
		return false
	}

	status := strings.Trim(string(body), "\n ")
	return status == "true"
}

//OpenRepository opens the repository at path. Currently
//verifies that it is a (bare) repository and returns an
//error if the check fails.
func OpenRepository(path string) (*Repository, error) {

	path, err := filepath.Abs(path)
	if err != nil {
		return nil, fmt.Errorf("git: could not determine absolute path")
	}

	if !IsBareRepository(path) {
		return nil, fmt.Errorf("git: not a bare repository")
	}

	return &Repository{Path: path}, nil
}

//DiscoverRepository returns the git repository that contains the
//current working directory, or and error if the current working
//dir does not lie inside one.
func DiscoverRepository() (*Repository, error) {
	cmd := exec.Command("git", "rev-parse", "--git-dir")
	data, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	path := strings.Trim(string(data), "\n ")
	return &Repository{Path: path}, nil
}

//ReadDescription returns the contents of the description file.
func (repo *Repository) ReadDescription() string {
	path := filepath.Join(repo.Path, "description")

	dat, err := ioutil.ReadFile(path)
	if err != nil {
		return ""
	}

	return string(dat)
}

//WriteDescription writes the contents of the description file.
func (repo *Repository) WriteDescription(description string) error {
	path := filepath.Join(repo.Path, "description")

	// not atomic, fine for now
	return ioutil.WriteFile(path, []byte(description), 0666)
}

// DeleteCollaborator removes a collaborator file from the repositories sharing folder.
func (repo *Repository) DeleteCollaborator(username string) error {
	filePath := filepath.Join(repo.Path, "gin", "sharing", username)

	return os.Remove(filePath)
}

//OpenObject returns the git object for a give id (SHA1).
func (repo *Repository) OpenObject(id SHA1) (Object, error) {
	obj, err := repo.openRawObject(id)

	if err != nil {
		return nil, err
	}

	if IsStandardObject(obj.otype) {
		return parseObject(obj)
	}

	//not a standard object, *must* be a delta object,
	// we know of no other types
	if !IsDeltaObject(obj.otype) {
		return nil, fmt.Errorf("git: unsupported object")
	}

	delta, err := parseDelta(obj)
	if err != nil {
		return nil, err
	}

	chain, err := buildDeltaChain(delta, repo)

	if err != nil {
		return nil, err
	}

	//TODO: check depth, and especially expected memory usage
	// beofre actually patching it

	return chain.resolve()
}

func (repo *Repository) openRawObject(id SHA1) (gitObject, error) {
	idstr := id.String()
	opath := filepath.Join(repo.Path, "objects", idstr[:2], idstr[2:])

	obj, err := openRawObject(opath)

	if err == nil {
		return obj, nil
	} else if err != nil && !os.IsNotExist(err) {
		return obj, err
	}

	indicies := repo.loadPackIndices()

	for _, f := range indicies {

		idx, err := PackIndexOpen(f)
		if err != nil {
			continue
		}

		//TODO: we should leave index files open,
		defer idx.Close()

		off, err := idx.FindOffset(id)

		if err != nil {
			continue
		}

		pf, err := idx.OpenPackFile()
		if err != nil {
			return gitObject{}, err
		}

		obj, err := pf.readRawObject(off)

		if err != nil {
			return gitObject{}, err
		}

		return obj, nil
	}

	// from inspecting the os.isNotExist source it
	// seems that if we have "not found" in the message
	// os.IsNotExist() report true, which is what we want
	return gitObject{}, fmt.Errorf("git: object not found")
}

func (repo *Repository) loadPackIndices() []string {
	target := filepath.Join(repo.Path, "objects", "pack", "*.idx")
	files, err := filepath.Glob(target)

	if err != nil {
		panic(err)
	}

	return files
}

//OpenRef returns the Ref with the given name or an error
//if either no maching could be found or in case the match
//was not unique.
func (repo *Repository) OpenRef(name string) (Ref, error) {

	if name == "HEAD" {
		return repo.parseRef("HEAD")
	}

	matches := repo.listRefWithName(name)

	//first search in local heads
	var locals []Ref
	for _, v := range matches {
		if IsBranchRef(v) {
			if name == v.Fullname() {
				return v, nil
			}
			locals = append(locals, v)
		}
	}

	// if we find a single local match
	// we return it directly
	if len(locals) == 1 {
		return locals[0], nil
	}

	switch len(matches) {
	case 0:
		return nil, fmt.Errorf("git: ref matching %q not found", name)
	case 1:
		return matches[0], nil
	}
	return nil, fmt.Errorf("git: ambiguous ref name, multiple matches")
}

//Readlink returns the destination of a symbilc link blob object
func (repo *Repository) Readlink(id SHA1) (string, error) {

	b, err := repo.OpenObject(id)
	if err != nil {
		return "", err
	}

	if b.Type() != ObjBlob {
		return "", fmt.Errorf("id must point to a blob")
	}

	blob := b.(*Blob)

	//TODO: check size and don't read unreasonable large blobs
	data, err := ioutil.ReadAll(blob)

	if err != nil {
		return "", err
	}

	return string(data), nil
}

//ObjectForPath will resolve the path to an object
//for the file tree starting in the node root.
//The root object can be either a Commit, Tree or Tag.
func (repo *Repository) ObjectForPath(root Object, pathstr string) (Object, error) {

	var node Object
	var err error

	switch o := root.(type) {
	case *Tree:
		node = root
	case *Commit:
		node, err = repo.OpenObject(o.Tree)
	case *Tag:
		node, err = repo.OpenObject(o.Object)
	default:
		return nil, fmt.Errorf("unsupported root object type")
	}

	if err != nil {
		return nil, fmt.Errorf("could not root tree object: %v", err)
	}

	cleaned := path.Clean(strings.Trim(pathstr, " /"))
	comps := strings.Split(cleaned, "/")

	var i int
	for i = 0; i < len(comps); i++ {

		tree, ok := node.(*Tree)
		if !ok {
			cwd := strings.Join(comps[:i+1], "/")
			err := &os.PathError{
				Op:   "convert git.Object to git.Tree",
				Path: cwd,
				Err:  fmt.Errorf("expected tree object, got %s", node.Type()),
			}
			return nil, err
		}

		//Since we call path.Clean(), this should really
		//only happen at the root, but it is safe to
		//have here anyway
		if comps[i] == "." || comps[i] == "/" {
			continue
		}

		var id *SHA1
		for tree.Next() {
			entry := tree.Entry()
			if entry.Name == comps[i] {
				id = &entry.ID
				break
			}
		}

		if err = tree.Err(); err != nil {
			cwd := strings.Join(comps[:i+1], "/")
			return nil, &os.PathError{
				Op:   "find object",
				Path: cwd,
				Err:  err}
		} else if id == nil {
			cwd := strings.Join(comps[:i+1], "/")
			return nil, &os.PathError{
				Op:   "find object",
				Path: cwd,
				Err:  os.ErrNotExist}
		}

		node, err = repo.OpenObject(*id)
		if err != nil {
			cwd := strings.Join(comps[:i+1], "/")
			return nil, &os.PathError{
				Op:   "open object",
				Path: cwd,
				Err:  err,
			}
		}
	}

	return node, nil
}

// usefmt is the option string used by CommitsForRef to return a formatted git commit log.
const usefmt = `--pretty=format:
Commit:=%H%n
Committer:=%cn%n
Author:=%an%n
Date-iso:=%ai%n
Date-rel:=%ar%n
Subject:=%s%n
Changes:=`

// CommitSummary represents a subset of information from a git commit.
type CommitSummary struct {
	Commit       string
	Committer    string
	Author       string
	DateIso      string
	DateRelative string
	Subject      string
	Changes      []string
}

// CommitsForRef executes a custom git log command for the specified ref of the
// associated git repository and returns the resulting byte array.
func (repo *Repository) CommitsForRef(ref string) ([]CommitSummary, error) {

	raw, err := commitsForRef(repo.Path, ref, usefmt)
	if err != nil {
		return nil, err
	}

	sep := ":="
	var comList []CommitSummary
	r := bytes.NewReader(raw)
	br := bufio.NewReader(r)

	var changesFlag bool
	for {
		// Consume line until newline character
		l, err := br.ReadString('\n')

		if strings.Contains(l, sep) {
			splitList := strings.SplitN(l, sep, 2)

			key := splitList[0]
			val := splitList[1]
			switch key {
			case "Commit":
				// reset non key line flags
				changesFlag = false
				newCommit := CommitSummary{Commit: val}
				comList = append(comList, newCommit)
			case "Committer":
				comList[len(comList)-1].Committer = val
			case "Author":
				comList[len(comList)-1].Author = val
			case "Date-iso":
				comList[len(comList)-1].DateIso = val
			case "Date-rel":
				comList[len(comList)-1].DateRelative = val
			case "Subject":
				comList[len(comList)-1].Subject = val
			case "Changes":
				// Setting changes flag so we know, that the next lines are probably file change notification lines.
				changesFlag = true
			default:
				fmt.Printf("[W] commits: unexpected key %q, value %q\n", key, strings.Trim(val, "\n"))
			}
		} else if changesFlag && strings.Contains(l, "\t") {
			comList[len(comList)-1].Changes = append(comList[len(comList)-1].Changes, l)
		}

		// Breaks at the latest when EOF err is raised
		if err != nil {
			break
		}
	}
	if err != io.EOF && err != nil {
		return nil, err
	}

	return comList, nil
}

// commitsForRef executes a custom git log command for the specified ref of the
// given git repository with the specified log format string and returns the resulting byte array.
// Function is kept private to force handling of the []byte inside the package.
func commitsForRef(repoPath, ref, usefmt string) ([]byte, error) {
	gdir := fmt.Sprintf("--git-dir=%s", repoPath)

	cmd := exec.Command("git", gdir, "log", ref, usefmt, "--name-status")
	body, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed running git log: %s\n", err.Error())
	}
	return body, nil
}

// BranchExists runs the "git branch <branchname> --list" command.
// It will return an error, if the command fails, true, if the result is not empty and false otherwise.
func (repo *Repository) BranchExists(branch string) (bool, error) {
	gdir := fmt.Sprintf("--git-dir=%s", repo.Path)

	cmd := exec.Command("git", gdir, "branch", branch, "--list")
	body, err := cmd.Output()
	if err != nil {
		return false, err
	} else if len(body) == 0 {
		return false, nil
	}

	return true, nil
}
