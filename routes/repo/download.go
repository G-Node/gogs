// Copyright 2014 The Gogs Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package repo

import (
	"io"
	"path"

	"github.com/gogits/git-module"

	"bufio"
	"github.com/G-Node/go-annex"
	"github.com/G-Node/gogs/pkg/context"
	"github.com/G-Node/gogs/pkg/setting"
	"github.com/G-Node/gogs/pkg/tool"
	"os"
	"github.com/go-macaron/captcha"
)

func ServeData(c *context.Context, name string, reader io.Reader, cpt *captcha.Captcha) error {
	buf := make([]byte, 1024)
	n, _ := reader.Read(buf)
	if n >= 0 {
		buf = buf[:n]
	}
	isannex := tool.IsAnnexedFile(buf)
	var afpR *bufio.Reader
	var afp *os.File
	if isannex == true {
		af, err := gannex.NewAFile(c.Repo.Repository.RepoPath(), "annex", name, buf)
		if err != nil {

		}
		if cpt!=nil && af.Info.Size() > gannex.MEGABYTE*setting.Repository.RawCaptchaMinFileSize && !cpt.VerifyReq(c.Req) &&
			!c.IsLogged {
			c.Data["EnableCaptcha"] = true
			c.HTML(200, "repo/download")
			return nil
		}
		afp, err = af.Open()
		defer afp.Close()
		if err != nil {

		}
		afpR = bufio.NewReader(afp)
		buf, _ = afpR.Peek(1024)
	}

	if !tool.IsTextFile(buf) {
		if !tool.IsImageFile(buf) {
			c.Resp.Header().Set("Content-Disposition", "attachment; filename=\""+name+"\"")
			c.Resp.Header().Set("Content-Transfer-Encoding", "binary")
		}
	} else if !setting.Repository.EnableRawFileRenderMode || !c.QueryBool("render") {
		c.Resp.Header().Set("Content-Type", "text/plain; charset=utf-8")
	}
	if isannex {
		_, err := io.Copy(c.Resp, afpR)
		return err
	}
	c.Resp.Write(buf)
	_, err := io.Copy(c.Resp, reader)
	return err
}

func ServeBlob(c *context.Context, blob *git.Blob, cpt *captcha.Captcha) error {
	r, w := io.Pipe()
	defer r.Close()
	defer w.Close()
	go blob.DataPipeline(w, w)
	return ServeData(c, path.Base(c.Repo.TreePath), io.LimitReader(r, blob.Size()), cpt)
}

func SingleDownload(c *context.Context, cpt *captcha.Captcha) {
	blob, err := c.Repo.Commit.GetBlobByPath(c.Repo.TreePath)
	if err != nil {
		if git.IsErrNotExist(err) {
			c.Handle(404, "GetBlobByPath", nil)
		} else {
			c.Handle(500, "GetBlobByPath", err)
		}
		return
	}
	// reallow direct download independent of size
	if err = ServeBlob(c, blob, nil); err != nil {
		c.Handle(500, "ServeBlob", err)
	}
}
