package dav

import (
	"github.com/G-Node/gogs/pkg/context"
	"golang.org/x/net/webdav"
	"os"
	"fmt"
)

func Dav(c *context.Context, handler *webdav.Handler) {
	handler.ServeHTTP(c.Resp, c.Req.Request)
	c.Write([]byte("Hallo"))
	return
}

// GinFS implements webdav (it implements webdav.Habdler) read only access to a repository
type GinFS struct{}

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
	return nil, nil
}

func (fs *GinFS) Stat(name string) (os.FileInfo, error) {
	return nil, nil
}
