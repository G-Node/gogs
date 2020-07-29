// Copyright 2014 The Gogs Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package admin

import (
	"github.com/unknwon/paginater"
	log "unknwon.dev/clog/v2"

	"github.com/G-Node/gogs/internal/conf"
	"github.com/G-Node/gogs/internal/context"
	"github.com/G-Node/gogs/internal/db"
)

const (
	REPOS = "admin/repo/list"
)

func Repos(c *context.Context) {
	c.Data["Title"] = c.Tr("admin.repositories")
	c.Data["PageIsAdmin"] = true
	c.Data["PageIsAdminRepositories"] = true

	page := c.QueryInt("page")
	if page <= 0 {
		page = 1
	}

	var (
		repos []*db.Repository
		count int64
		err   error
	)

	keyword := c.Query("q")
	if len(keyword) == 0 {
		repos, err = db.Repositories(page, conf.UI.Admin.RepoPagingNum)
		if err != nil {
			c.Handle(500, "Repositories", err)
			return
		}
		count = db.CountRepositories(true)
	} else {
		repos, count, err = db.SearchRepositoryByName(&db.SearchRepoOptions{
			Keyword:  keyword,
			OrderBy:  "id ASC",
			Private:  true,
			Page:     page,
			PageSize: conf.UI.Admin.RepoPagingNum,
		})
		if err != nil {
			c.Handle(500, "SearchRepositoryByName", err)
			return
		}
	}
	c.Data["Keyword"] = keyword
	c.Data["Total"] = count
	c.Data["Page"] = paginater.New(int(count), conf.UI.Admin.RepoPagingNum, page, 5)

	if err = db.RepositoryList(repos).LoadAttributes(); err != nil {
		c.Handle(500, "LoadAttributes", err)
		return
	}
	c.Data["Repos"] = repos

	c.HTML(200, REPOS)
}

func DeleteRepo(c *context.Context) {
	repo, err := db.GetRepositoryByID(c.QueryInt64("id"))
	if err != nil {
		c.Handle(500, "GetRepositoryByID", err)
		return
	}

	if err := db.DeleteRepository(repo.MustOwner().ID, repo.ID); err != nil {
		c.Handle(500, "DeleteRepository", err)
		return
	}
	log.Trace("Repository deleted: %s/%s", repo.MustOwner().Name, repo.Name)

	c.Flash.Success(c.Tr("repo.settings.deletion_success"))
	c.JSON(200, map[string]interface{}{
		"redirect": conf.Server.Subpath + "/admin/repos?page=" + c.Query("page"),
	})
}
