package repo

import (
	"bufio"
	"bytes"
	"io"
	"io/ioutil"

	"github.com/G-Node/git-module"
	gannex "github.com/G-Node/go-annex"
	"github.com/G-Node/gogs/pkg/context"
	"github.com/G-Node/gogs/pkg/setting"
	"github.com/G-Node/gogs/pkg/tool"
	"github.com/G-Node/libgin/libgin"
	"github.com/go-macaron/captcha"
	log "gopkg.in/clog.v1"
	"gopkg.in/yaml.v2"
)

func serveAnnexedData(ctx *context.Context, name string, cpt *captcha.Captcha, buf []byte) error {
	annexFile, err := gannex.NewAFile(ctx.Repo.Repository.RepoPath(), "annex", name, buf)
	if err != nil {
		return err
	}
	if cpt != nil && annexFile.Info.Size() > gannex.MEGABYTE*setting.Repository.RawCaptchaMinFileSize && !cpt.VerifyReq(ctx.Req) &&
		!ctx.IsLogged {
		ctx.Data["EnableCaptcha"] = true
		ctx.HTML(200, "repo/download")
		return nil
	}
	annexfp, err := annexFile.Open()
	if err != nil {
		return err
	}
	defer annexfp.Close()
	annexReader := bufio.NewReader(annexfp)
	buf, _ = annexReader.Peek(1024)

	if !tool.IsTextFile(buf) {
		if !tool.IsImageFile(buf) {
			ctx.Resp.Header().Set("Content-Disposition", "attachment; filename=\""+name+"\"")
			ctx.Resp.Header().Set("Content-Transfer-Encoding", "binary")
		}
	} else if !setting.Repository.EnableRawFileRenderMode || !ctx.QueryBool("render") {
		ctx.Resp.Header().Set("Content-Type", "text/plain; charset=utf-8")
	}

	_, err = io.Copy(ctx.Resp, annexReader)
	return err
}

func readDataciteFile(entry *git.TreeEntry, c *context.Context) {
	c.Data["HasDatacite"] = true
	doiData, err := entry.Blob().Data()
	if err != nil {
		log.Trace("DOI Blob could not be read: %v", err)
	}
	buf, err := ioutil.ReadAll(doiData)
	doiInfo := libgin.DOIRegInfo{}
	err = yaml.Unmarshal(buf, &doiInfo)
	if err != nil {
		log.Trace("DOI Blob could not be unmarshalled: %v", err)
	}
	c.Data["DOIInfo"] = &doiInfo

	doi := calcRepoDOI(c, setting.DOI.DOIBase)
	//ddata, err := ginDoi.GDoiMData(doi, "https://api.datacite.org/works/") //todo configure URL?

	c.Data["DOIReg"] = libgin.IsRegisteredDOI(doi)
	c.Data["DOI"] = doi
}

// calcRepoDOI calculates the theoretical DOI for a repository. If the repository
// belongs to the DOI user (and is a fork) it uses the name for the base
// repository.
func calcRepoDOI(c *context.Context, doiBase string) string {
	repoN := c.Repo.Repository.FullName()
	// check whether this repo belongs to DOI and is a fork
	if c.Repo.Repository.IsFork && c.Repo.Owner.Name == "doi" {
		repoN = c.Repo.Repository.BaseRepo.FullName()
	}
	uuid := libgin.RepoPathToUUID(repoN)
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
func resolveAnnexedContent(c *context.Context, buf []byte, dataRc io.Reader) ([]byte, io.Reader) {
	if !tool.IsAnnexedFile(buf) {
		// not an annex pointer file; return as is
		return buf, dataRc
	}
	log.Trace("Annexed file requested: Resolving content for [%s]", bytes.TrimSpace(buf))
	af, err := gannex.NewAFile(c.Repo.Repository.RepoPath(), "annex", "", buf)
	if err != nil {
		log.Trace("Could not get annex file: %v", err)
		c.ServerError("readmeFile.Data", err)
		return buf, dataRc
	}

	afp, err := af.Open()
	if err != nil {
		c.ServerError("readmeFile.Data", err)
		log.Trace("Could not open annex file: %v", err)
		return buf, dataRc
	}
	annexDataReader := bufio.NewReader(afp)
	annexBuf := make([]byte, 1024)
	n, _ := annexDataReader.Read(annexBuf)
	annexBuf = annexBuf[:n]
	c.Data["FileSize"] = af.Info.Size()
	log.Trace("Annexed file size: %d B", af.Info.Size())
	return annexBuf, annexDataReader
}
