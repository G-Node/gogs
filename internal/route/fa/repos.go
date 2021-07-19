// Copyright 2014 The Gogs Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package fa

import (
	"github.com/unknwon/paginater"

	"github.com/ivis-yoshida/gogs/internal/conf"
	"github.com/ivis-yoshida/gogs/internal/context"
	"github.com/ivis-yoshida/gogs/internal/db"
)

const (
	REPOS = "fa/repo/list"
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
			c.Error(err, "list repositories")
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
			c.Error(err, "search repository by name")
			return
		}
	}
	c.Data["Keyword"] = keyword
	c.Data["Total"] = count
	c.Data["Page"] = paginater.New(int(count), conf.UI.Admin.RepoPagingNum, page, 5)

	if err = db.RepositoryList(repos).LoadAttributes(); err != nil {
		c.Error(err, "load attributes")
		return
	}
	c.Data["Repos"] = repos

	c.Success(REPOS)
}
