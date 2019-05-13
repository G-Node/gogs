package repo

import (
	"bufio"
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
