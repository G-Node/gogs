package repo

import (
	"bufio"
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/gogs/git-module"
	"github.com/ivis-yoshida/gogs/internal/conf"
	"github.com/ivis-yoshida/gogs/internal/context"
	"github.com/ivis-yoshida/gogs/internal/db"
	"github.com/ivis-yoshida/gogs/internal/tool"
	log "gopkg.in/clog.v1"
)

func serveAnnexedData(ctx *context.Context, name string, buf []byte) error {
	keyparts := strings.Split(strings.TrimSpace(string(buf)), "/")
	key := keyparts[len(keyparts)-1]
	contentPath, err := git.NewCommand("annex", "contentlocation", key).RunInDir(ctx.Repo.Repository.RepoPath())
	if err != nil {
		log.Error(2, "Failed to find content location for file %q with key %q", name, key)
		return err
	}
	// always trim space from output for git command
	contentPath = bytes.TrimSpace(contentPath)
	return serveAnnexedKey(ctx, name, string(contentPath))
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
	} else if !conf.Repository.EnableRawFileRenderMode || !ctx.QueryBool("render") {
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

// readDmpJson is RCOS specific code.
func readDmpJson(c *context.Context) {
	log.Trace("Reading dmp.json file")
	entry, err := c.Repo.Commit.Blob("/dmp.json")
	if err != nil || entry == nil {
		log.Error(2, "dmp.json blob could not be retrieved: %v", err)
		c.Data["HasDmpJson"] = false
		return
	}
	buf, err := entry.Bytes()
	if err != nil {
		log.Error(2, "dmp.json data could not be read: %v", err)
		c.Data["HasDmpJson"] = false
		return
	}
	c.Data["DOIInfo"] = string(buf)
}

// GenerateMaDmp is RCOS specific code.
// This generates maDMP(machine actionable DMP) based on
// DMP information created by the user in the repository.
func GenerateMaDmp(c *context.Context) {
	// GitHubテンプレートNotebookを取得
	// refs: 1. https://zenn.dev/snowcait/scraps/3d51d8f7841f0c
	//       2. https://qiita.com/taizo/items/c397dbfed7215969b0a5
	template_url := "https://api.github.com/repos/ivis-kuwata/maDMP-template/contents/maDMP.ipynb"
	contents, err := fetchBlobOnGithub(template_url)
	if err != nil {
		log.Error(2, "maDMP blob could not be retrieved: %v", err)

		failedGenereteMaDmp(c, "Sorry, faild gerate maDMP: fetching template failed")
		return
	}

	var blob interface{}
	err = json.Unmarshal(contents, &blob)
	if err != nil {
		log.Error(2, "maDMP blob could not be retrieved: %v", err)

		failedGenereteMaDmp(c, "Sorry, faild gerate maDMP: unmarshal template failed")
		return
	}

	raw := blob.(map[string]interface{})["content"]
	decodedMaDmp, err := base64.StdEncoding.DecodeString(raw.(string))
	if err != nil {
		log.Error(2, "maDMP blob could not be retrieved: %v", err)

		failedGenereteMaDmp(c, "Sorry, faild gerate maDMP: decode template failed")
		return
	}

	// ユーザが作成したDMP情報取得
	entry, err := c.Repo.Commit.Blob("/dmp.json")
	if err != nil || entry == nil {
		log.Error(2, "dmp.json blob could not be retrieved: %v", err)

		failedGenereteMaDmp(c, "Sorry, faild gerate maDMP: DMP could not read")
		return
	}
	buf, err := entry.Bytes()
	if err != nil {
		log.Error(2, "dmp.json data could not be read: %v", err)

		failedGenereteMaDmp(c, "Sorry, faild gerate maDMP: DMP could not read")
		return
	}

	var dmp interface{}
	err = json.Unmarshal(buf, &dmp)
	if err != nil {
		log.Error(2, "Unmarshal DMP info: %v", err)

		failedGenereteMaDmp(c, "Sorry, faild gerate maDMP: DMP could not read")
		return
	}

	// dmp.jsonに"fields"プロパティがある想定
	selectedField := dmp.(map[string]interface{})["field"]
	/* maDMPへ埋め込む情報を追加する際は
	ここに追記のこと
	e.g.
	hasGrdm := dmp.(map[string]interface{})["hasGrdm"]
	*/

	pathToMaDmp := "maDMP.ipynb"
	err = c.Repo.Repository.UpdateRepoFile(c.User, db.UpdateRepoFileOptions{
		LastCommitID: c.Repo.CommitID,
		OldBranch:    c.Repo.BranchName,
		NewBranch:    c.Repo.BranchName,
		OldTreeName:  "",
		NewTreeName:  pathToMaDmp,
		Message:      "[GIN] Generate maDMP",
		Content: fmt.Sprintf(
			string(decodedMaDmp), // 埋め込み先: maDMP
			selectedField,        // 埋め込む値: DMP情報(現在は"field"の値のみ)
			/* maDMPへ埋め込む情報を追加する際は
			ここに追記のこと
			e.g.
			hasGrdm, */
		),
		IsNewFile: true,
	})
	if err != nil {
		log.Error(2, "failed generating maDMP: %v", err)

		failedGenereteMaDmp(c, "Faild gerate maDMP: Already exist")
		return
	}

	c.Flash.Success("maDMP generated!")
	c.Redirect(c.Repo.RepoLink)
}

// fetchBlobOnGithub is RCOS specific code.
// This uses the Github API to retrieve information about the file
// specified in the argument, and returns it in the type of []byte.
// If any processing fails, it will return error.
func fetchBlobOnGithub(blobPath string) ([]byte, error) {
	req, err := http.NewRequest("GET", blobPath, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	client := new(http.Client)

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	contents, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return contents, nil
}

// failedGenerateMaDmp is RCOS specific code.
// This is a function used by GenerateMaDmp to emit an error message
// on UI when maDMP generation fails.
func failedGenereteMaDmp(c *context.Context, msg string) {
	c.Flash.Error(msg)
	c.Redirect(c.Repo.RepoLink)
}

// resolveAnnexedContent takes a buffer with the contents of a git-annex
// pointer file and an io.Reader for the underlying file and returns the
// corresponding buffer and a bufio.Reader for the underlying content file.
// The returned byte slice and bufio.Reader can be used to replace the buffer
// and io.Reader sent in through the caller so that any existing code can use
// the two variables without modifications.
// Any errors that occur during processing are stored in the provided context.
// The FileSize of the annexed content is also saved in the context (c.Data["FileSize"]).
func resolveAnnexedContent(c *context.Context, buf []byte) ([]byte, error) {
	if !tool.IsAnnexedFile(buf) {
		// not an annex pointer file; return as is
		return buf, nil
	}
	log.Trace("Annexed file requested: Resolving content for %q", bytes.TrimSpace(buf))

	keyparts := strings.Split(strings.TrimSpace(string(buf)), "/")
	key := keyparts[len(keyparts)-1]
	contentPath, err := git.NewCommand("annex", "contentlocation", key).RunInDir(c.Repo.Repository.RepoPath())
	if err != nil {
		log.Error(2, "Failed to find content location for key %q", key)
		c.Data["IsAnnexedFile"] = true
		return buf, err
	}
	// always trim space from output for git command
	contentPath = bytes.TrimSpace(contentPath)
	afp, err := os.Open(filepath.Join(c.Repo.Repository.RepoPath(), string(contentPath)))
	if err != nil {
		log.Trace("Could not open annex file: %v", err)
		c.Data["IsAnnexedFile"] = true
		return buf, err
	}
	info, err := afp.Stat()
	if err != nil {
		log.Trace("Could not stat annex file: %v", err)
		c.Data["IsAnnexedFile"] = true
		return buf, err
	}
	annexDataReader := bufio.NewReader(afp)
	annexBuf := make([]byte, 1024)
	n, _ := annexDataReader.Read(annexBuf)
	annexBuf = annexBuf[:n]
	c.Data["FileSize"] = info.Size()
	log.Trace("Annexed file size: %d B", info.Size())
	return annexBuf, nil
}

func GitConfig(c *context.Context) {
	log.Trace("RepoPath: %+v", c.Repo.Repository.RepoPath())
	configFilePath := path.Join(c.Repo.Repository.RepoPath(), "config")
	log.Trace("Serving file %q", configFilePath)
	if _, err := os.Stat(configFilePath); err != nil {
		c.Error(err, "GitConfig")
		// c.ServerError("GitConfig", err)
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
		c.Error(err, "AnnexGetKey")
	}
}
