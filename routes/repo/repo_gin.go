package repo

import (
	"bufio"
	"io"

	gannex "github.com/G-Node/go-annex"
	"github.com/G-Node/gogs/pkg/context"
	"github.com/G-Node/gogs/pkg/setting"
	"github.com/G-Node/gogs/pkg/tool"
	"github.com/go-macaron/captcha"
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
