// Copyright 2014 The Gogs Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package repo

import (
	"fmt"
	"net/http"
	"path"

	"github.com/gogs/git-module"

	"github.com/NII-DG/gogs/internal/conf"
	"github.com/NII-DG/gogs/internal/context"
	"github.com/NII-DG/gogs/internal/tool"
	logv2 "unknwon.dev/clog/v2"
)

func serveData(c *context.Context, name string, data []byte) error {
	commit, err := c.Repo.Commit.CommitByPath(git.CommitByRevisionOptions{Path: c.Repo.TreePath})
	if err != nil {
		return fmt.Errorf("get commit by path %q: %v", c.Repo.TreePath, err)
	}
	c.Resp.Header().Set("Last-Modified", commit.Committer.When.Format(http.TimeFormat))

	if tool.IsAnnexedFile(data) {
		return serveAnnexedData(c, name, data)
	}

	if !tool.IsTextFile(data) {
		if !tool.IsImageFile(data) {
			c.Resp.Header().Set("Content-Disposition", "attachment; filename=\""+name+"\"")
			c.Resp.Header().Set("Content-Transfer-Encoding", "binary")
		}
	} else if !conf.Repository.EnableRawFileRenderMode || !c.QueryBool("render") {
		c.Resp.Header().Set("Content-Type", "text/plain; charset=utf-8")
	}

	if _, err := c.Resp.Write(data); err != nil {
		return fmt.Errorf("write buffer to response: %v", err)
	}
	return nil
}

func ServeBlob(c *context.Context, blob *git.Blob) error {
	p, err := blob.Bytes()
	if err != nil {
		return err
	}

	return serveData(c, path.Base(c.Repo.TreePath), p)
}

func SingleDownload(c *context.Context) {
	logv2.Info("c.Repo.TreePath", c.Repo.TreePath)
	blob, err := c.Repo.Commit.Blob(c.Repo.TreePath)
	tree, terr := c.Repo.Commit.TreeEntry(c.Repo.TreePath)
	if terr != nil {
		logv2.Error("terr : %v", terr)
	} else {
		logv2.Info("tree.IsBlob() : %v", tree.IsBlob())
		logv2.Info("tree.IsExec() : %v", tree.IsExec())
		logv2.Info("tree.Mode() : %v", tree.Mode())
		logv2.Info("tree.Name() : %v", tree.Name())
		logv2.Info("tree.Type() : %v", tree.Type())
		logv2.Info("tree.Size() : %v", tree.Size())
		logv2.Info("tree.Size() : %v", tree.Size())
		b, e := tree.Blob().Bytes()
		if e != nil {
			logv2.Error("tree.Blob().Bytes() : ERR %v", e)
		}
		logv2.Info("tree.Blob().Bytes() : %v", string(b))
	}

	entries, ierr := c.Repo.Commit.Tree.Entries()
	if ierr != nil {
		logv2.Error("ierr : %v", ierr)
	} else {
		for _, v := range entries {
			logv2.Info("v : %v", v.Name())
		}
	}
	t, terr := c.Repo.Commit.Subtree("test_data")
	if ierr != nil {
		logv2.Error("ierr : %v", ierr)
	} else {

		entries, ierr = t.Entries()
		if ierr != nil {
			logv2.Error("ierr : %v", ierr)
		} else {
			for _, v := range entries {
				logv2.Info("v2 : %v", v.Name())
			}
		}
	}

	// if err != nil {
	// 	logv2.Error("Repo.Commit.Blob() ERR : %v", err)
	// 	c.NotFoundOrError(gitutil.NewError(err), "get blob")
	// 	return
	// }

	if err = ServeBlob(c, blob); err != nil {
		c.Error(err, "serve blob")
		return
	}
}
