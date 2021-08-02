// Copyright 2014 The Gogs Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package route

import (
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"strconv"

	"github.com/go-macaron/i18n"
	"github.com/gogs/git-module"
	"github.com/unknwon/paginater"
	log "gopkg.in/clog.v1"
	"gopkg.in/macaron.v1"

	"github.com/ivis-yoshida/gogs/internal/conf"
	"github.com/ivis-yoshida/gogs/internal/context"
	"github.com/ivis-yoshida/gogs/internal/db"
	"github.com/ivis-yoshida/gogs/internal/route/repo/dmp_schema"
	"github.com/ivis-yoshida/gogs/internal/route/user"
)

const (
	HOME                  = "home"
	EXPLORE_REPOS         = "explore/repos"
	EXPLORE_USERS         = "explore/users"
	EXPLORE_ORGANIZATIONS = "explore/organizations"
	EXPLORE_METADATA      = "explore/metadata"
)

func Home(c *context.Context) {
	if c.IsLogged {
		if !c.User.IsActive && conf.Auth.RequireEmailConfirmation {
			c.Data["Title"] = c.Tr("auth.active_your_account")
			c.Success(user.ACTIVATE)
		} else {
			user.Dashboard(c)
		}
		return
	}

	// Check auto-login.
	uname := c.GetCookie(conf.Security.CookieUsername)
	if len(uname) != 0 {
		c.Redirect(conf.Server.Subpath + "/user/login")
		return
	}

	c.Data["PageIsHome"] = true
	c.Success(HOME)
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
	repos, count, err := db.SearchRepositoryByName(&db.SearchRepoOptions{
		Keyword:  keyword,
		UserID:   c.UserID(),
		OrderBy:  "updated_unix DESC",
		Page:     page,
		PageSize: conf.UI.ExplorePagingNum,
	})
	if err != nil {
		c.Error(err, "search repository by name")
		return
	}
	c.Data["Keyword"] = keyword
	c.Data["Total"] = count
	c.Data["Page"] = paginater.New(int(count), conf.UI.ExplorePagingNum, page, 5)

	if err = db.RepositoryList(repos).LoadAttributes(); err != nil {
		c.Error(err, "load attributes")
		return
	}
	c.Data["Repos"] = filterUnlistedRepos(repos)

	c.Success(EXPLORE_REPOS)
}

// ExploreMetadata is RCOS specific code
func ExploreMetadata(c *context.Context) {
	c.Data["Title"] = c.Tr("explore")
	c.Data["PageIsExplore"] = true
	c.Data["PageIsExploreMetadata"] = true

	page := c.QueryInt("page")
	if page <= 0 {
		page = 1
	}

	// fetch query parameter
	keyword := c.Query("q")
	selectedKey := c.Query("selectKey")
	repos, count, err := db.SearchRepositoryByName(&db.SearchRepoOptions{
		Keyword:  "",
		UserID:   c.UserID(),
		OrderBy:  "updated_unix DESC",
		Page:     page,
		PageSize: conf.UI.ExplorePagingNum,
	})
	if err != nil {
		c.Error(err, "search repository by name")
		return
	}

	// 	Get dmp.json contents
	// for _, repo := range repos {
	repos[0].HasMetadata = false
	gitRepo, repoErr := git.Open(repos[0].RepoPath())
	if repoErr != nil {
		c.Error(repoErr, "open repository")
		return
	}
	commit, err := gitRepo.CatFileCommit("refs/heads/master")
	entry, err := commit.Blob("/dmp.json")
	if err != nil || entry == nil {
		log.Error(2, "datacite.yml blob could not be retrieved: %v", err)
		c.Data["HasDataCite"] = false
		return
	}
	buf, err := entry.Bytes()
	if err != nil {
		log.Error(2, "datacite.yml data could not be read: %v", err)
		c.Data["HasDataCite"] = false
		return
	}

	// FIXME : multiple schema
	dmpContents := dmp_schema.MetiDmpInfo{}

	err = json.Unmarshal(buf, &dmpContents)
	if err != nil {
		log.Error(2, "dmp.json data could not be unmarshalled: %v", err)
		c.Data["HasDataCite"] = false
		return
	}

	c.Data["DOIInfo"] = &dmpContents

	if selectedKey != "" && keyword != "" && isContained(dmpContents, selectedKey, keyword) {
		c.Data["SelectedKey"] = selectedKey
		c.Data["SearchResult"] = keyword
		repos[0].HasMetadata = true
		// c.Data["HasContent"] = true

		c.Data["TmpString"] = isContained(dmpContents, selectedKey, keyword)
	}
	// }
	// below is search

	c.Data["Keyword"] = keyword
	c.Data["Total"] = count
	c.Data["Page"] = paginater.New(int(count), conf.UI.ExplorePagingNum, page, 5)

	if err = db.RepositoryList(repos).LoadAttributes(); err != nil {
		c.Error(err, "load attributes")
		return
	}
	c.Data["Repos"] = filterUnlistedRepos(repos)

	c.Success(EXPLORE_METADATA)
}

func isContained(srt dmp_schema.MetiDmpInfo, selectedIndex, keyword string) bool {
	srtValue := reflect.ValueOf(srt)
	i, err := strconv.Atoi(selectedIndex)
	if err != nil {
		log.Error(2, "selectedIndex can't parse int: %v", err)
	}
	v := srtValue.Field(i)

	return v.String() == keyword
}

type UserSearchOptions struct {
	Type     db.UserType
	Counter  func() int64
	Ranger   func(int, int) ([]*db.User, error)
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
		users []*db.User
		count int64
		err   error
	)

	keyword := c.Query("q")
	if len(keyword) == 0 {
		users, err = opts.Ranger(page, opts.PageSize)
		if err != nil {
			c.Error(err, "ranger")
			return
		}
		count = opts.Counter()
	} else {
		users, count, err = db.SearchUserByName(&db.SearchUserOptions{
			Keyword:  keyword,
			Type:     opts.Type,
			OrderBy:  opts.OrderBy,
			Page:     page,
			PageSize: opts.PageSize,
		})
		if err != nil {
			c.Error(err, "search user by name")
			return
		}
	}
	c.Data["Keyword"] = keyword
	c.Data["Total"] = count
	c.Data["Page"] = paginater.New(int(count), opts.PageSize, page, 5)
	c.Data["Users"] = users

	c.Success(opts.TplName)
}

func ExploreUsers(c *context.Context) {
	c.Data["Title"] = c.Tr("explore")
	c.Data["PageIsExplore"] = true
	c.Data["PageIsExploreUsers"] = true

	RenderUserSearch(c, &UserSearchOptions{
		Type:     db.UserIndividual,
		Counter:  db.CountUsers,
		Ranger:   db.ListUsers,
		PageSize: conf.UI.ExplorePagingNum,
		OrderBy:  "updated_unix DESC",
		TplName:  EXPLORE_USERS,
	})
}

func ExploreOrganizations(c *context.Context) {
	c.Data["Title"] = c.Tr("explore")
	c.Data["PageIsExplore"] = true
	c.Data["PageIsExploreOrganizations"] = true

	RenderUserSearch(c, &UserSearchOptions{
		Type:     db.UserOrganization,
		Counter:  db.CountOrganizations,
		Ranger:   db.Organizations,
		PageSize: conf.UI.ExplorePagingNum,
		OrderBy:  "updated_unix DESC",
		TplName:  EXPLORE_ORGANIZATIONS,
	})
}

func NotFound(c *macaron.Context, l i18n.Locale) {
	c.Data["Title"] = l.Tr("status.page_not_found")
	c.HTML(http.StatusNotFound, fmt.Sprintf("status/%d", http.StatusNotFound))
}
