package dav

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	"io/ioutil"

	"github.com/G-Node/git-module"
	gctx "github.com/G-Node/gogs/internal/context"
	"github.com/G-Node/gogs/internal/db"
	"github.com/G-Node/gogs/internal/setting"
	"github.com/G-Node/gogs/internal/tool"
	"github.com/G-Node/libgin/libgin/annex"
	"golang.org/x/net/context"
	"golang.org/x/net/webdav"
	log "gopkg.in/clog.v1"
)

var (
	RE_GETRNAME = regexp.MustCompile(`.+\/(.+)\/_dav`)
	RE_GETROWN  = regexp.MustCompile(`\/(.+)\/.+\/_dav`)
	RE_GETFPATH = regexp.MustCompile("/_dav/(.+)")
)

const ANNEXPEEKSIZE = 1024

func Dav(c *gctx.Context, handler *webdav.Handler) {
	if !setting.WebDav.On {
		c.WriteHeader(http.StatusUnauthorized)
		return
	}
	if checkPerms(c) != nil {
		Webdav401(c)
		return
	}
	handler.ServeHTTP(c.Resp, c.Req.Request)
	return
}

// GinFS implements webdav (it implements webdav.Habdler) read only access to a repository
type GinFS struct {
	BasePath string
}

// Mkdir not implemented: Just return an error. -> Read Only
func (fs *GinFS) Mkdir(ctx context.Context, name string, perm os.FileMode) error {
	return fmt.Errorf("Mkdir not implemented for read only gin FS")
}

// RemoveAll not implemented: Just return an error. -> Read Only
func (fs *GinFS) RemoveAll(ctx context.Context, name string) error {
	return fmt.Errorf("RemoveAll not implemented for read only gin FS")
}

// Rename not implemented: Just return an error. -> Read Only
func (fs *GinFS) Rename(ctx context.Context, oldName, newName string) error {
	return fmt.Errorf("Rename not implemented for read only gin FS")
}

// OpenFile returns a named file from the repository
func (fs *GinFS) OpenFile(ctx context.Context, name string, flag int, perm os.FileMode) (webdav.File, error) {
	//todo: catch all the errors
	rname, err := getRName(name)
	if err != nil {
		return nil, err
	}

	oname, err := getOName(name)
	if err != nil {
		return nil, err
	}

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
	return &GinFile{trentry: trentry, tree: tree, LChange: com.Committer.When, rpath: rpath}, nil
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
	rpath     string
	afp       *os.File
}

func (f *GinFile) Write(p []byte) (n int, err error) {
	return 0, fmt.Errorf("write to GinFile not implemented (read only)")
}

func (f *GinFile) Close() error {
	if f.afp != nil {
		return f.afp.Close()
	}
	return nil
}

func (f *GinFile) read(p []byte) (int, error) {
	if f.trentry == nil {
		return 0, fmt.Errorf("File not found")
	}
	if f.trentry.Type != git.OBJECT_BLOB {
		return 0, fmt.Errorf("not a blob")
	}
	data, err := f.trentry.Blob().Data()
	if err != nil {
		return 0, err
	}
	// todo: read with pipes
	io.CopyN(ioutil.Discard, data, f.seekoset)
	n, err := data.Read(p)
	if err != nil {
		return n, err
	}
	return n, nil
}
func (f *GinFile) Read(p []byte) (int, error) {
	if f.afp != nil {
		return f.afp.Read(p)
	}
	tmp := make([]byte, len(p))
	n, err := f.read(tmp)
	tmp = tmp[:n]
	if err != nil {
		return n, err
	}

	annexed := tool.IsAnnexedFile(tmp)
	if annexed {
		af, err := annex.NewAFile(f.rpath, "annex", f.trentry.Name(), tmp)
		if err != nil {
			return n, err
		}
		f.afp, _ = af.Open()
		f.afp.Seek(f.seekoset, io.SeekStart)
		return f.afp.Read(p)
	}
	copy(p, tmp)
	f.Seek(int64(n), io.SeekCurrent)
	return n, nil
}

func (f *GinFile) Seek(offset int64, whence int) (int64, error) {
	if f.afp != nil {
		return f.afp.Seek(offset, whence)
	}
	st, err := f.Stat()
	if err != nil {
		return f.seekoset, err
	}
	switch whence {
	case io.SeekStart:
		if offset > st.Size() || offset < 0 {
			return 0, fmt.Errorf("cannot seek to %d, only %d big", offset, st.Size())
		}
		f.seekoset = offset
		return f.seekoset, nil
	case io.SeekCurrent:
		noffset := f.seekoset + offset
		if noffset > st.Size() || noffset < 0 {
			return 0, fmt.Errorf("cannot seek to %d, only %d big", offset, st.Size())
		}
		f.seekoset = noffset
		return f.seekoset, nil
	case io.SeekEnd:
		fsize := st.Size()
		noffset := fsize - offset
		if noffset > fsize || noffset < 0 {
			return 0, fmt.Errorf("cannot seek to %d, only %d big", offset, st.Size())
		}
		f.seekoset = noffset
		return f.seekoset, nil
	}
	return f.seekoset, fmt.Errorf("seeking failed")
}

func (f *GinFile) Readdir(count int) ([]os.FileInfo, error) {
	ents, err := f.tree.ListEntries()
	if err != nil {
		return nil, err
	}
	// give back all the stuff
	if count <= 0 {
		return f.getFInfos(ents)
	}
	// user requested a bufferrd read
	switch {
	case count > len(ents):
		infos, err := f.getFInfos(ents)
		if err != nil {
			return nil, err
		}
		return infos, io.EOF
	case f.dirrcount >= len(ents):
		return nil, io.EOF
	case f.dirrcount+count >= len(ents):
		infos, err := f.getFInfos(ents[f.dirrcount:])
		if err != nil {
			return nil, err
		}
		f.dirrcount = len(ents)
		return infos, io.EOF
	case f.dirrcount+count < len(ents):
		infos, err := f.getFInfos(ents[f.dirrcount : f.dirrcount+count])
		if err != nil {
			return nil, err
		}
		f.dirrcount = f.dirrcount + count
		return infos, nil
	}
	return nil, nil
}

func (f *GinFile) getFInfos(ents []*git.TreeEntry) ([]os.FileInfo, error) {
	infos := make([]os.FileInfo, len(ents))
	for c, ent := range ents {
		finfo, err := GinFile{trentry: ent, rpath: f.rpath}.Stat()
		if err != nil {
			return nil, err
		}
		infos[c] = finfo
	}
	return infos, nil
}
func (f GinFile) Stat() (os.FileInfo, error) {
	if f.trentry == nil {
		return nil, fmt.Errorf("File not found")
	}
	if f.trentry.Type != git.OBJECT_BLOB {
		return GinFinfo{TreeEntry: f.trentry, LChange: f.LChange}, nil
	}
	peek := make([]byte, ANNEXPEEKSIZE)
	offset := f.seekoset
	f.seekoset = 0
	n, err := f.read(peek)
	f.seekoset = offset
	if err != nil {
		return nil, err
	}
	peek = peek[:n]
	if tool.IsAnnexedFile(peek) {
		af, err := annex.NewAFile(f.rpath, "annex", f.trentry.Name(), peek)
		if err != nil {
			return nil, err
		}
		f.trentry.SetSize(af.Info.Size())
	}
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
	if !c.Repo.HasAccess() {
		return fmt.Errorf("no access")
	}
	if !setting.WebDav.Logged {
		return nil
	}
	if !c.IsLogged {
		return fmt.Errorf("no access")
	}
	return nil
}

func getRepo(path string) (*db.Repository, error) {
	oID, err := getROwnerID(path)
	if err != nil {
		return nil, err
	}

	rname, err := getRName(path)
	if err != nil {
		return nil, err
	}

	return db.GetRepositoryByName(oID, rname)
}

func getRName(path string) (string, error) {
	name := RE_GETRNAME.FindStringSubmatch(path)
	if len(name) > 1 {
		return strings.ToLower(name[1]), nil
	}
	return "", fmt.Errorf("could not determine repo name")
}

func getOName(path string) (string, error) {
	name := RE_GETROWN.FindStringSubmatch(path)
	if len(name) > 1 {
		return strings.ToLower(name[1]), nil
	}
	return "", fmt.Errorf("could not determine repo owner")
}

func getFPath(path string) (string, error) {
	name := RE_GETFPATH.FindStringSubmatch(path)
	if len(name) > 1 {
		return name[1], nil
	}
	return "", fmt.Errorf("could not determine file path from %s", name)
}

func getROwnerID(path string) (int64, error) {
	name := RE_GETROWN.FindStringSubmatch(path)
	if len(name) > 1 {
		db.GetUserByName(name[1])
	}
	return -100, fmt.Errorf("could not determine repo owner")
}

func Webdav401(c *gctx.Context) {
	c.Header().Add("WWW-Authenticate", fmt.Sprintf("Basic realm=\"%s\"", setting.WebDav.AuthRealm))
	c.WriteHeader(http.StatusUnauthorized)
	return
}

func Logger(req *http.Request, err error) {
	if err != nil {
		log.Info("davlog: err:%+v", err)
		log.Trace("davlog: req:%+v", req)
	}
}
