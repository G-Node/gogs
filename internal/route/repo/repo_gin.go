package repo

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/G-Node/git-module"
	"github.com/G-Node/gogs/internal/context"
	"github.com/G-Node/gogs/internal/setting"
	"github.com/G-Node/gogs/internal/tool"
	"github.com/G-Node/libgin/libgin"
	"github.com/go-macaron/captcha"
	log "gopkg.in/clog.v1"
	"gopkg.in/yaml.v2"
)

func serveAnnexedData(ctx *context.Context, name string, cpt *captcha.Captcha, buf []byte) error {
	keyparts := strings.Split(strings.TrimSpace(string(buf)), "/")
	key := keyparts[len(keyparts)-1]
	contentPath, err := git.NewCommand("annex", "contentlocation", key).RunInDir(ctx.Repo.Repository.RepoPath())
	if err != nil {
		log.Error(2, "Failed to find content location for file %q with key %q", name, key)
		return err
	}
	// always trim space from output for git command
	contentPath = strings.TrimSpace(contentPath)
	return serveAnnexedKey(ctx, name, contentPath)
}

func serveAnnexedKey(ctx *context.Context, name string, contentPath string) error {
	fullContentPath := filepath.Join(ctx.Repo.Repository.RepoPath(), contentPath)
	annexfp, err := os.Open(fullContentPath)
	if err != nil {
		log.Error(2, "Failed to open annex file at %q: %s", fullContentPath, err.Error())
		return err
	}
	defer annexfp.Close()
	annexReader := bufio.NewReader(annexfp)

	info, err := annexfp.Stat()
	if err != nil {
		log.Error(2, "Failed to stat file at %q: %s", fullContentPath, err.Error())
		return err
	}

	buf, _ := annexReader.Peek(1024)

	ctx.Resp.Header().Set("Content-Length", fmt.Sprintf("%d", info.Size()))
	if !tool.IsTextFile(buf) {
		if !tool.IsImageFile(buf) {
			ctx.Resp.Header().Set("Content-Disposition", "attachment; filename=\""+name+"\"")
			ctx.Resp.Header().Set("Content-Transfer-Encoding", "binary")
		}
	} else if !setting.Repository.EnableRawFileRenderMode || !ctx.QueryBool("render") {
		ctx.Resp.Header().Set("Content-Type", "text/plain; charset=utf-8")
	}

	log.Trace("Serving annex content for %q: %q", name, contentPath)
	if ctx.Req.Method == http.MethodHead {
		// Skip content copy when request method is HEAD
		log.Trace("Returning header: %+v", ctx.Resp.Header())
		return nil
	}
	_, err = io.Copy(ctx.Resp, annexReader)
	return err
}

func readDataciteFile(entry *git.Blob, c *context.Context) {
	log.Trace("Found datacite.yml file")
	c.Data["HasDataCite"] = true
	doiData, err := entry.Data()
	if err != nil {
		log.Error(2, "datacite.yml blob could not be read: %v", err)
		return
	}
	buf, err := ioutil.ReadAll(doiData)
	if err != nil {
		log.Error(2, "datacite.yml data could not be read: %v", err)
		return
	}
	doiInfo := libgin.DOIRegInfo{}
	err = yaml.Unmarshal(buf, &doiInfo)
	if err != nil {
		log.Error(2, "datacite.yml data could not be unmarshalled: %v", err)
		return
	}
	c.Data["DOIInfo"] = &doiInfo
	c.Data["IsDOIReady"] = isDOIReady(c)
}

// True if repository is not Private, is not registered, or is registered and
// has changes.
func isDOIReady(c *context.Context) bool {
	if hasDC, ok := c.Data["HasDataCite"]; !ok || !hasDC.(bool) {
		return false
	}
	dbrepo := c.Repo.Repository
	gitrepo := c.Repo.GitRepo

	headIsRegistered := func() bool {
		currentDOI, ok := c.Data["DOI"]
		if !ok {
			return false
		}

		headBranch, err := gitrepo.GetHEADBranch()
		if err != nil {
			log.Error(2, "Failed to get HEAD branch for repo at %q: %v", gitrepo.Path, err)
			return false
		}

		headCommit, err := gitrepo.GetBranchCommitID(headBranch.Name)
		if err != nil {
			log.Error(2, "Failed to get commit ID of branch %q for repo at %q: %v", headBranch.Name, gitrepo.Path, err)
		}

		// if current valid and registered DOI matches the HEAD commit, can't
		// register again
		doiCommit, err := gitrepo.GetTagCommitID(currentDOI.(string))
		if err != nil {
			log.Error(2, "Failed to get commit ID of tag %q for repo at %q: %v", currentDOI, gitrepo.Path, err)
		}

		log.Trace("%s ?= %s", headCommit, doiCommit)
		return headCommit == doiCommit
	}()

	return !dbrepo.IsPrivate && !headIsRegistered
}

// resolveAnnexedContent takes a buffer with the contents of a git-annex
// pointer file and an io.Reader for the underlying file and returns the
// corresponding buffer and a bufio.Reader for the underlying content file.
// The returned byte slice and bufio.Reader can be used to replace the buffer
// and io.Reader sent in through the caller so that any existing code can use
// the two variables without modifications.
// Any errors that occur during processing are stored in the provided context.
// The FileSize of the annexed content is also saved in the context (c.Data["FileSize"]).
func resolveAnnexedContent(c *context.Context, buf []byte, dataRc io.Reader) ([]byte, io.Reader, error) {
	if !tool.IsAnnexedFile(buf) {
		// not an annex pointer file; return as is
		return buf, dataRc, nil
	}
	log.Trace("Annexed file requested: Resolving content for %q", bytes.TrimSpace(buf))

	keyparts := strings.Split(strings.TrimSpace(string(buf)), "/")
	key := keyparts[len(keyparts)-1]
	contentPath, err := git.NewCommand("annex", "contentlocation", key).RunInDir(c.Repo.Repository.RepoPath())
	if err != nil {
		log.Error(2, "Failed to find content location for key %q", key)
		c.Data["IsAnnexedFile"] = true
		return buf, dataRc, err
	}
	// always trim space from output for git command
	contentPath = strings.TrimSpace(contentPath)
	afp, err := os.Open(filepath.Join(c.Repo.Repository.RepoPath(), contentPath))
	if err != nil {
		log.Trace("Could not open annex file: %v", err)
		c.Data["IsAnnexedFile"] = true
		return buf, dataRc, err
	}
	info, err := afp.Stat()
	if err != nil {
		log.Trace("Could not stat annex file: %v", err)
		c.Data["IsAnnexedFile"] = true
		return buf, dataRc, err
	}
	annexDataReader := bufio.NewReader(afp)
	annexBuf := make([]byte, 1024)
	n, _ := annexDataReader.Read(annexBuf)
	annexBuf = annexBuf[:n]
	c.Data["FileSize"] = info.Size()
	log.Trace("Annexed file size: %d B", info.Size())
	return annexBuf, annexDataReader, nil
}

func GitConfig(c *context.Context) {
	log.Trace("RepoPath: %+v", c.Repo.Repository.RepoPath())
	configFilePath := path.Join(c.Repo.Repository.RepoPath(), "config")
	log.Trace("Serving file %q", configFilePath)
	if _, err := os.Stat(configFilePath); err != nil {
		c.ServerError("GitConfig", err)
		return
	}
	c.ServeFileContent(configFilePath, "config")
}

func AnnexGetKey(c *context.Context) {
	filename := c.Params(":keyfile")
	key := c.Params(":key")
	contentPath := filepath.Join("annex/objects", c.Params(":hashdira"), c.Params(":hashdirb"), key, filename)
	log.Trace("Git annex requested key %q: %q", key, contentPath)
	err := serveAnnexedKey(c, filename, contentPath)
	if err != nil {
		c.ServerError("AnnexGetKey", err)
	}
}
