// Copyright 2014 The Gogs Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package repo

import (
	"io"
	"path"

	"github.com/gogs/git-module"

	"github.com/gogs/gogs/pkg/context"
	"github.com/gogs/gogs/pkg/setting"
	"github.com/gogs/gogs/pkg/tool"
)

func ServeData(c *context.Context, name string, reader io.Reader) error {
	buf := make([]byte, 1024)
	n, _ := reader.Read(buf)
	if n >= 0 {
		buf = buf[:n]
	}

	if !tool.IsTextFile(buf) {
		if !tool.IsImageFile(buf) {
			c.Resp.Header().Set("Content-Disposition", "attachment; filename=\""+name+"\"")
			c.Resp.Header().Set("Content-Transfer-Encoding", "binary")
		}
	} else if !setting.Repository.EnableRawFileRenderMode || !c.QueryBool("render") {
		c.Resp.Header().Set("Content-Type", "text/plain; charset=utf-8")
	}
	c.Resp.Write(buf)
	_, err := io.Copy(c.Resp, reader)
	return err
}

func ServeBlob(c *context.Context, blob *git.Blob) error {
	r, w := io.Pipe()
	defer r.Close()
	defer w.Close()
	go blob.DataPipeline(w, w)
	return ServeData(c, path.Base(c.Repo.TreePath), io.LimitReader(r, blob.Size()))
}

func SingleDownload(c *context.Context) {
	blob, err := c.Repo.Commit.GetBlobByPath(c.Repo.TreePath)
	if err != nil {
		if git.IsErrNotExist(err) {
			c.Handle(404, "GetBlobByPath", nil)
		} else {
			c.Handle(500, "GetBlobByPath", err)
		}
		return
	}
	if err = ServeBlob(c, blob); err != nil {
		c.Handle(500, "ServeBlob", err)
	}
}
