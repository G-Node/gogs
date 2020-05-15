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
	"github.com/G-Node/gogs/internal/db"
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

func readDataciteFile(entry *git.TreeEntry, c *context.Context) {
	log.Trace("Found datacite.yml file")
	c.Data["HasDatacite"] = true
	doiData, err := entry.Blob().Data()
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

	if doi := getRepoDOI(c); doi != "" && libgin.IsRegisteredDOI(doi) {
		c.Data["DOI"] = doi
	}
}

// getRepoDOI returns the DOI for the repository based on the following rules:
// - if the repository belongs to the DOI user and has a tag that matches the
// DOI prefix, returns the tag.
// - if the repo is forked by the DOI user, check the DOI fork for the tag as above.
// - if the repo is forked by the DOI user and the fork doesn't have a tag,
// returns the (old-style) calculated DOI, based on the hash of the repository
// path.
// - An empty string is returned if it is not not forked by the DOI user.
// If an error occurs at any point, returns an empty string (the error is logged).
// Tag retrieval is allowed to fail and falls back on the hashed DOI method.
func getRepoDOI(c *context.Context) string {
	repo := c.Repo.Repository
	var doiFork *db.Repository
	if repo.Owner.Name == "doi" {
		doiFork = repo
	} else {
		if forks, err := repo.GetForks(); err == nil {
			for _, fork := range forks {
				if fork.MustOwner().Name == "doi" {
					doiFork = fork
					break
				}
			}
		} else {
			log.Error(2, "failed to get forks for repository %q (%d): %v", repo.FullName(), repo.ID, err)
			return ""
		}
	}

	if doiFork == nil {
		// not owned or forked by DOI, so not registered
		return ""
	}

	// check the DOI fork for a tag that matches our DOI prefix
	// if multiple exit, get the latest one
	doiBase := setting.DOI.Base

	doiForkGit, err := git.OpenRepository(doiFork.RepoPath())
	if err != nil {
		log.Error(2, "failed to open git repository at %q (%d): %v", doiFork.RepoPath(), doiFork.ID, err)
		return ""
	}
	if tags, err := doiForkGit.GetTags(); err == nil {
		var latestTime int64
		latestTag := ""
		for _, tagName := range tags {
			if strings.Contains(tagName, doiBase) {
				tag, err := doiForkGit.GetTag(tagName)
				if err != nil {
					// log the error and continue to the next tag
					log.Error(2, "failed to get information for tag %q for repository at %q: %v", tagName, doiForkGit.Path, err)
					continue
				}
				commit, err := tag.Commit()
				if err != nil {
					// log the error and continue to the next tag
					log.Error(2, "failed to get commit for tag %q for repository at %q: %v", tagName, doiForkGit.Path, err)
					continue
				}
				commitTime := commit.Committer.When.Unix()
				if commitTime > latestTime {
					latestTag = tagName
					latestTime = commitTime
				}
				return latestTag
			}
		}
	} else {
		// this shouldn't happen even if there are no tags
		// log the error, but fall back to the old method anyway
		log.Error(2, "failed to get tags for repository at %q: %v", doiForkGit.Path, err)
	}

	// Has DOI fork but isn't tagged: return old style has-based DOI
	repoPath := repo.FullName()
	// get base repo name if it's a DOI fork
	if c.Repo.Repository.IsFork && c.Repo.Owner.Name == "doi" {
		repoPath = c.Repo.Repository.BaseRepo.FullName()
	}
	uuid := libgin.RepoPathToUUID(repoPath)
	return doiBase + uuid[:6]
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
