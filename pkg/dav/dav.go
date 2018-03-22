package dav

import (
	"github.com/G-Node/gogs/pkg/context"
	"golang.org/x/net/webdav"
	"os"
	"fmt"
	"net/http"
	"github.com/G-Node/gogs/models"
	"regexp"
	"github.com/G-Node/git-module"
	"time"
)

var (
	RE_GETRNAME = regexp.MustCompile(".+/(.+)/_dav")
	RE_GETROWN  = regexp.MustCompile("./(.+)/.+/_dav")
	RE_GETFPATH = regexp.MustCompile("/_dav/(.+)")
)

func Dav(c *context.Context, handler *webdav.Handler) {
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
func (fs *GinFS) Mkdir(name string, perm os.FileMode) error {
	return fmt.Errorf("Mkdir not implemented for read only gin FS")
}

// Just return an error. -> Read Only
func (fs *GinFS) RemoveAll(name string) error {
	return fmt.Errorf("Remove not implemented for read only gin FS")
}

// Just return an error. -> Read Only
func (fs *GinFS) Rename(oldName, newName string) error {
	return fmt.Errorf("Rename not implemented for read only gin FS")
}

func (fs *GinFS) OpenFile(name string, flag int, perm os.FileMode) (webdav.File, error) {
	//todo: catch all the errors
	rname, _ := getRName(name)
	oname, _ := getOName(name)
	path, _ := getFPath(name)
	grepo, _ := git.OpenRepository(fmt.Sprintf("%s/%s/%s.git", oname, rname))
	com, _ := grepo.GetBranchCommit("master")
	ent, _ := com.GetTreeEntryByPath(path)
	return GinFile{Trentry: ent}, nil
}

func (fs GinFS) Stat(name string) (os.FileInfo, error) {
	return nil, nil
}

type GinFile struct {
	Trentry *git.TreeEntry
}

func (f GinFile) Write(p []byte) (n int, err error) {
	return 0, fmt.Errorf("Write to GinFile not implemented (read only)")
}

func (f GinFile) Close() error {
	return nil
}

func (f GinFile) Read(p []byte) (n int, err error) {
	return 0, nil
}

func (f GinFile) Seek(offset int64, whence int) (int64, error) {
	return 0, nil
}

func (f GinFile) Readdir(count int) ([]os.FileInfo, error) {
	return nil, nil
}

func (f GinFile) Stat() (os.FileInfo, error) {
	return GinFinfo{f.Trentry}, nil
}

type GinFinfo struct {
	*git.TreeEntry
}

func (i GinFinfo) Mode() os.FileMode {
	return 0
}

func (i GinFinfo) ModTime() time.Time {
	return time.Now()
}

func (i GinFinfo) Sys() interface{} {
	return nil
}


func checkPerms(c *context.Context) error {
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
