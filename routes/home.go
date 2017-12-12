// Copyright 2014 The Gogs Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package routes

import (
	"github.com/Unknwon/paginater"

	"github.com/G-Node/gogs/models"
	"github.com/G-Node/gogs/pkg/context"
	"github.com/G-Node/gogs/pkg/setting"
	"github.com/G-Node/gogs/routes/user"
	"strings"
)

const (
	HOME                  = "home"
	EXPLORE_REPOS         = "explore/repos"
	EXPLORE_USERS         = "explore/users"
	EXPLORE_ORGANIZATIONS = "explore/organizations"
)

func Home(c *context.Context) {
	if c.IsLogged {
		if !c.User.IsActive && setting.Service.RegisterEmailConfirm {
			c.Data["Title"] = c.Tr("auth.active_your_account")
			c.HTML(200, user.ACTIVATE)
		} else {
			user.Dashboard(c)
		}
		return
	}

	// Check auto-login.
	uname := c.GetCookie(setting.CookieUserName)
	if len(uname) != 0 {
		c.Redirect(setting.AppSubURL + "/user/login")
		return
	}

	c.Data["PageIsHome"] = true
	c.HTML(200, HOME)
}

func ExploreRepos(c *context.Context) {
	c.Data["Title"] = c.Tr("explore")
	c.Data["PageIsExplore"] = true
	c.Data["PageIsExploreRepositories"] = true

	page := c.QueryInt("page")
	if page <= 0 {
		page = 1
	}

	keyword := c.Query("q")
	repos, count, err := models.SearchRepositoryByName(&models.SearchRepoOptions{
		Keyword:  keyword,
		UserID:   c.UserID(),
		OrderBy:  "updated_unix DESC",
		Page:     page,
		PageSize: setting.UI.ExplorePagingNum,
	})
	if err != nil {
		c.Handle(500, "SearchRepositoryByName", err)
		return
	}
	c.Data["Keyword"] = keyword
	c.Data["Total"] = count
	c.Data["Page"] = paginater.New(int(count), setting.UI.ExplorePagingNum, page, 5)

	if err = models.RepositoryList(repos).LoadAttributes(); err != nil {
		c.Handle(500, "LoadAttributes", err)
		return
	}
	// filter repos we eant to not show in list
	var showRep []*models.Repository
	for _, repo := range repos {
		if !strings.Contains(repo.Name, "hideme") && !strings.Contains(repo.Name, "unlisted") {
			showRep = append(showRep, repo)
		}
	}

	c.Data["Repos"] = showRep

	c.HTML(200, EXPLORE_REPOS)
}

type UserSearchOptions struct {
	Type     models.UserType
	Counter  func() int64
	Ranger   func(int, int) ([]*models.User, error)
	PageSize int
	OrderBy  string
	TplName  string
}

func RenderUserSearch(c *context.Context, opts *UserSearchOptions) {
	page := c.QueryInt("page")
	if page <= 1 {
		page = 1
	}

	var (
		users []*models.User
		count int64
		err   error
	)

	keyword := c.Query("q")
	if len(keyword) == 0 {
		users, err = opts.Ranger(page, opts.PageSize)
		if err != nil {
			c.Handle(500, "opts.Ranger", err)
			return
		}
		count = opts.Counter()
	} else {
		users, count, err = models.SearchUserByName(&models.SearchUserOptions{
			Keyword:  keyword,
			Type:     opts.Type,
			OrderBy:  opts.OrderBy,
			Page:     page,
			PageSize: opts.PageSize,
		})
		if err != nil {
			c.Handle(500, "SearchUserByName", err)
			return
		}
	}
	c.Data["Keyword"] = keyword
	c.Data["Total"] = count
	c.Data["Page"] = paginater.New(int(count), opts.PageSize, page, 5)
	c.Data["Users"] = users

	c.HTML(200, opts.TplName)
}

func ExploreUsers(c *context.Context) {
	c.Data["Title"] = c.Tr("explore")
	c.Data["PageIsExplore"] = true
	c.Data["PageIsExploreUsers"] = true

	RenderUserSearch(c, &UserSearchOptions{
		Type:     models.USER_TYPE_INDIVIDUAL,
		Counter:  models.CountUsers,
		Ranger:   models.Users,
		PageSize: setting.UI.ExplorePagingNum,
		OrderBy:  "updated_unix DESC",
		TplName:  EXPLORE_USERS,
	})
}

func ExploreOrganizations(c *context.Context) {
	c.Data["Title"] = c.Tr("explore")
	c.Data["PageIsExplore"] = true
	c.Data["PageIsExploreOrganizations"] = true

	RenderUserSearch(c, &UserSearchOptions{
		Type:     models.USER_TYPE_ORGANIZATION,
		Counter:  models.CountOrganizations,
		Ranger:   models.Organizations,
		PageSize: setting.UI.ExplorePagingNum,
		OrderBy:  "updated_unix DESC",
		TplName:  EXPLORE_ORGANIZATIONS,
	})
}

func NotFound(c *context.Context) {
	c.Data["Title"] = "Page Not Found"
	c.NotFound()
}
