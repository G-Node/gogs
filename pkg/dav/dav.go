package dav

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"time"

	"github.com/G-Node/git-module"
	"github.com/G-Node/gogs/models"
	gctx "github.com/G-Node/gogs/pkg/context"
	"golang.org/x/net/context"
	"golang.org/x/net/webdav"
)

var (
	RE_GETRNAME = regexp.MustCompile(`.+\/(.+)\/_dav`)
	RE_GETROWN  = regexp.MustCompile(`\/(.+)\/.+\/_dav`)
	RE_GETFPATH = regexp.MustCompile("/_dav/(.+)")
)

func Dav(c *gctx.Context, handler *webdav.Handler) {
	if checkPerms(c) != nil {
		c.WriteHeader(http.StatusUnauthorized)
		return
	}
	handler.ServeHTTP(c.Resp, c.Req.Request)
	return
}

// GinFS implements webdav (it implements webdav.Habdler) read only access to a repository
type GinFS struct {
	BasePath string
}

// Just return an error. -> Read Only
func (fs *GinFS) Mkdir(ctx context.Context, name string, perm os.FileMode) error {
	return fmt.Errorf("Mkdir not implemented for read only gin FS")
}

// Just return an error. -> Read Only
func (fs *GinFS) RemoveAll(ctx context.Context, name string) error {
	return fmt.Errorf("Remove not implemented for read only gin FS")
}

// Just return an error. -> Read Only
func (fs *GinFS) Rename(ctx context.Context, oldName, newName string) error {
	return fmt.Errorf("Rename not implemented for read only gin FS")
}

func (fs *GinFS) OpenFile(ctx context.Context, name string, flag int, perm os.FileMode) (webdav.File, error) {
	//todo: catch all the errors
	rname, _ := getRName(name)
	oname, _ := getOName(name)
	path, _ := getFPath(name)
	rpath := fmt.Sprintf("%s/%s/%s.git", fs.BasePath, oname, rname)
	grepo, err := git.OpenRepository(rpath)
	if err != nil {
		return nil, err
	}
	com, err := grepo.GetBranchCommit("master")
	if err != nil {
		return nil, err
	}
	tree, _ := com.SubTree(path)
	trentry, _ := com.GetTreeEntryByPath(path)
	return &GinFile{trentry: trentry, tree: tree, LChange: com.Committer.When}, nil
}

func (fs *GinFS) Stat(ctx context.Context, name string) (os.FileInfo, error) {
	f, err := fs.OpenFile(ctx, name, 0, 0)
	if err != nil {
		return nil, err
	}
	return f.Stat()
}

type GinFile struct {
	tree      *git.Tree
	trentry   *git.TreeEntry
	dirrcount int
	seekoset  int64
	LChange   time.Time
}

func (f *GinFile) Write(p []byte) (n int, err error) {
	return 0, fmt.Errorf("Write to GinFile not implemented (read only)")
}

func (f *GinFile) Close() error {
	return nil
}

func (f *GinFile) Read(p []byte) (n int, err error) {
	if f.trentry.Type != git.OBJECT_BLOB {
		return 0, fmt.Errorf("not a blob")
	}
	data, err := f.trentry.Blob().Data()
	if err != nil {
		return 0, err
	}
	// todo: annex
	return data.Read(p)
}

func (f *GinFile) Seek(offset int64, whence int) (int64, error) {
	switch whence {
	case 0:
		f.seekoset = offset
	case 1:
		f.seekoset = f.seekoset + offset
	case 2:
		fstat, err := f.Stat()
		if err != nil {
			return -1, err
		}
		fsize := fstat.Size()
		f.seekoset = fsize - offset
	}
	return 0, nil
}

func (f *GinFile) Readdir(count int) ([]os.FileInfo, error) {
	ents, err := f.tree.ListEntries()
	if err != nil {
		return nil, err
	}
	// give back all the stuff
	if count <= 0 {
		return getFInfos(ents)
	}
	// user requested a bufferrd read
	switch {
	case count > len(ents):
		infos, err := getFInfos(ents)
		if err != nil {
			return nil, err
		}
		return infos, io.EOF
	case f.dirrcount >= len(ents):
		return nil, io.EOF
	case f.dirrcount+count >= len(ents):
		infos, err := getFInfos(ents[f.dirrcount:])
		if err != nil {
			return nil, err
		}
		f.dirrcount = len(ents)
		return infos, io.EOF
	case f.dirrcount+count < len(ents):
		infos, err := getFInfos(ents[f.dirrcount : f.dirrcount+count])
		if err != nil {
			return nil, err
		}
		f.dirrcount = f.dirrcount + count
		return infos, nil
	}
	return nil, nil
}

func getFInfos(ents []*git.TreeEntry) ([]os.FileInfo, error) {
	infos := make([]os.FileInfo, len(ents))
	for c, ent := range ents {
		finfo, err := GinFile{trentry: ent}.Stat()
		if err != nil {
			return nil, err
		}
		infos[c] = finfo
	}
	return infos, nil
}
func (f GinFile) Stat() (os.FileInfo, error) {
	return GinFinfo{TreeEntry: f.trentry, LChange: f.LChange}, nil
}

type GinFinfo struct {
	*git.TreeEntry
	LChange time.Time
}

func (i GinFinfo) Mode() os.FileMode {
	return 0
}

func (i GinFinfo) ModTime() time.Time {
	return i.LChange
}

func (i GinFinfo) Sys() interface{} {
	return nil
}

func checkPerms(c *gctx.Context) error {
	return nil
}

func getRepo(path string) (*models.Repository, error) {
	oID, err := getROwnerID(path)
	if err != nil {
		return nil, err
	}

	rname, err := getRName(path)
	if err != nil {
		return nil, err
	}

	return models.GetRepositoryByName(oID, rname)
}

func getRName(path string) (string, error) {
	name := RE_GETRNAME.FindStringSubmatch(path)
	if len(name) > 1 {
		return name[1], nil
	}
	return "", fmt.Errorf("Could not determine repo name")
}

func getOName(path string) (string, error) {
	name := RE_GETROWN.FindStringSubmatch(path)
	if len(name) > 1 {
		return name[1], nil
	}
	return "", fmt.Errorf("Could not determine repo owner")
}

func getFPath(path string) (string, error) {
	name := RE_GETFPATH.FindStringSubmatch(path)
	if len(name) > 1 {
		return name[1], nil
	}
	return "", fmt.Errorf("Could not determine file path")
}

func getROwnerID(path string) (int64, error) {
	name := RE_GETROWN.FindStringSubmatch(path)
	if len(name) > 1 {
		models.GetUserByName(name[1])
	}
	return -100, fmt.Errorf("Could not determine repo owner")
}
