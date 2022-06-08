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

	"github.com/NII-DG/gogs/internal/conf"
	"github.com/NII-DG/gogs/internal/context"
	"github.com/NII-DG/gogs/internal/db"
	"github.com/NII-DG/gogs/internal/tool"
	"github.com/gogs/git-module"
	log "gopkg.in/clog.v1"
	logv2 "unknwon.dev/clog/v2"
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
	logv2.Trace("fullContentPath : %v", fullContentPath)
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
func readDmpJson(c context.AbstructContext) {
	log.Trace("Reading dmp.json file")
	entry, err := c.GetRepo().GetCommit().Blob("/dmp.json")
	if err != nil || entry == nil {
		log.Error(2, "dmp.json blob could not be retrieved: %v", err)
		c.CallData()["HasDmpJson"] = false
		return
	}
	buf, err := entry.Bytes()
	if err != nil {
		log.Error(2, "dmp.json data could not be read: %v", err)
		c.CallData()["HasDmpJson"] = false
		return
	}
	c.CallData()["DOIInfo"] = string(buf)
}

// GenerateMaDmp is RCOS specific code.
func GenerateMaDmp(c context.AbstructContext) {
	var f repoUtil
	generateMaDmp(c, f)
}

// generateMaDmp is RCOS specific code.
// This generates maDMP(machine actionable DMP) based on
// DMP information created by the user in the repository.
func generateMaDmp(c context.AbstructContext, f AbstructRepoUtil) {
	// GitHubテンプレートNotebookを取得
	// refs: 1. https://zenn.dev/snowcait/scraps/3d51d8f7841f0c
	//       2. https://qiita.com/taizo/items/c397dbfed7215969b0a5
	templateUrl := getTemplateUrl() + "maDMP.ipynb"

	src, err := f.FetchContentsOnGithub(templateUrl)
	if err != nil {
		log.Error(2, "maDMP blob could not be fetched: %v", err)
	}

	decodedMaDmp, err := f.DecodeBlobContent(src)
	if err != nil {
		log.Error(2, "maDMP blob could not be decorded: %v", err)

		failedGenereteMaDmp(c, "Sorry, faild gerate maDMP: fetching template failed")
		return
	}

	// コード付帯機能の起動時間短縮のための暫定的な定義
	fetchDockerfile(c)

	// ユーザが作成したDMP情報取得
	entry, err := c.GetRepo().GetCommit().Blob("/dmp.json")
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
	selectedDataSize := dmp.(map[string]interface{})["dataSize"]
	selectedDatasetStructure := dmp.(map[string]interface{})["datasetStructure"]
	/* maDMPへ埋め込む情報を追加する際は
	ここに追記のこと
	e.g.
	hasGrdm := dmp.(map[string]interface{})["hasGrdm"]
	*/

	pathToMaDmp := "maDMP.ipynb"
	err = c.GetRepo().GetDbRepo().UpdateRepoFile(c.GetUser(), db.UpdateRepoFileOptions{
		LastCommitID: c.GetRepo().GetLastCommitIdStr(),
		OldBranch:    c.GetRepo().GetBranchName(),
		NewBranch:    c.GetRepo().GetBranchName(),
		OldTreeName:  "",
		NewTreeName:  pathToMaDmp,
		Message:      "[GIN] Generate maDMP",
		Content: fmt.Sprintf(
			decodedMaDmp,  // この行が埋め込み先: maDMP
			selectedField, // ここより以下は埋め込む値: DMP情報
			selectedDataSize,
			selectedDatasetStructure,
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

	c.GetFlash().Success("maDMP generated!")
	c.Redirect(c.GetRepo().GetRepoLink())
}

type AbstructRepoUtil interface {
	FetchContentsOnGithub(blobPath string) ([]byte, error)
	DecodeBlobContent(blobInfo []byte) (string, error)
}

type repoUtil func()

func (f repoUtil) FetchContentsOnGithub(blobPath string) ([]byte, error) {
	return f.fetchContentsOnGithub(blobPath)
}

func (f repoUtil) DecodeBlobContent(blobInfo []byte) (string, error) {
	return f.decodeBlobContent(blobInfo)
}

// FetchContentsOnGithub is RCOS specific code.
// This uses the Github API to retrieve information about the file
// specified in the argument, and returns it in the type of []byte.
// If any processing fails, it will return error.
// refs: https://docs.github.com/en/rest/reference/repos#contents
func (f repoUtil) fetchContentsOnGithub(blobPath string) ([]byte, error) {
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
	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("Error: blob not found.")
	}
	defer resp.Body.Close()

	contents, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return contents, nil
}

// DecodeBlobContent is RCOS specific code.
// This reads and decodes "content" value of the response byte slice
// retrieved from the GitHub API.
// refs: https://docs.github.com/en/rest/reference/repos#contents
func (f repoUtil) decodeBlobContent(blobInfo []byte) (string, error) {
	var blob interface{}
	err := json.Unmarshal(blobInfo, &blob)
	if err != nil {
		return "", err
	}

	raw := blob.(map[string]interface{})["content"]
	decodedBlobContent, err := base64.StdEncoding.DecodeString(raw.(string))
	if err != nil {
		return "", err
	}

	return string(decodedBlobContent), nil
}

// failedGenerateMaDmp is RCOS specific code.
// This is a function used by GenerateMaDmp to emit an error message
// on UI when maDMP generation fails.
func failedGenereteMaDmp(c context.AbstructContext, msg string) {
	c.GetFlash().Error(msg)
	c.Redirect(c.GetRepo().GetRepoLink())
}

// fetchDockerfile is RCOS specific code.
// This fetches the Dockerfile used when launching Binderhub.
func fetchDockerfile(c context.AbstructContext) {
	// コード付帯機能の起動時間短縮のための暫定的な定義
	dockerfileUrl := getTemplateUrl() + "Dockerfile"

	var f repoUtil
	src, err := f.FetchContentsOnGithub(dockerfileUrl)
	if err != nil {
		log.Error(2, "Dockerfile could not be fetched: %v", err)
	}

	decodedDockerfile, err := f.DecodeBlobContent(src)
	if err != nil {
		log.Error(2, "Dockerfile could not be decorded: %v", err)

		failedGenereteMaDmp(c, "Sorry, faild gerate maDMP: fetching template failed(Dockerfile)")
		return
	}

	pathToDockerfile := "Dockerfile"
	_ = c.GetRepo().GetDbRepo().UpdateRepoFile(c.GetUser(), db.UpdateRepoFileOptions{
		LastCommitID: c.GetRepo().GetLastCommitIdStr(),
		OldBranch:    c.GetRepo().GetBranchName(),
		NewBranch:    c.GetRepo().GetBranchName(),
		OldTreeName:  "",
		NewTreeName:  pathToDockerfile,
		Message:      "[GIN] fetch Dockerfile",
		Content:      decodedDockerfile,
		IsNewFile:    true,
	})
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
	log.Info("Annexed file requested: Resolving content for %q", bytes.TrimSpace(buf))

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
