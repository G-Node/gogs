// Copyright 2014 The Gogs Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package repo

import (
	"fmt"
	"io"
	"net/http"
	"path"

	"github.com/G-Node/git-module"
	"github.com/go-macaron/captcha"

	"github.com/G-Node/gogs/internal/conf"
	"github.com/G-Node/gogs/internal/context"
	"github.com/G-Node/gogs/internal/tool"
)

func serveData(c *context.Context, name string, r io.Reader, cpt *captcha.Captcha) error {
	buf := make([]byte, 1024)
	n, _ := r.Read(buf)
	if n >= 0 {
		buf = buf[:n]
	}

	commit, err := c.Repo.Commit.GetCommitByPath(c.Repo.TreePath)
	if err != nil {
		return fmt.Errorf("GetCommitByPath: %v", err)
	}
	c.Resp.Header().Set("Last-Modified", commit.Committer.When.Format(http.TimeFormat))

	if tool.IsAnnexedFile(buf) {
		return serveAnnexedData(c, name, cpt, buf)
	}

	if !tool.IsTextFile(buf) {
		if !tool.IsImageFile(buf) {
			c.Resp.Header().Set("Content-Disposition", "attachment; filename=\""+name+"\"")
			c.Resp.Header().Set("Content-Transfer-Encoding", "binary")
		}
	} else if !conf.Repository.EnableRawFileRenderMode || !c.QueryBool("render") {
		c.Resp.Header().Set("Content-Type", "text/plain; charset=utf-8")
	}

	if _, err := c.Resp.Write(buf); err != nil {
		return fmt.Errorf("write buffer to response: %v", err)
	}
	_, err = io.Copy(c.Resp, r)
	return err
}

func ServeBlob(c *context.Context, blob *git.Blob, cpt *captcha.Captcha) error {
	r, w := io.Pipe()
	defer r.Close()
	defer w.Close()
	go blob.DataPipeline(w, w)
	return serveData(c, path.Base(c.Repo.TreePath), io.LimitReader(r, blob.Size()), cpt)
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
