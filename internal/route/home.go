// Copyright 2014 The Gogs Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package route

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/go-macaron/i18n"
	"github.com/gogs/git-module"
	"github.com/unknwon/paginater"
	log "gopkg.in/clog.v1"
	"gopkg.in/macaron.v1"

	"github.com/NII-DG/gogs/internal/conf"
	"github.com/NII-DG/gogs/internal/context"
	"github.com/NII-DG/gogs/internal/db"
	"github.com/NII-DG/gogs/internal/route/user"
)

const (
	HOME                  = "home"
	EXPLORE_REPOS         = "explore/repos"
	EXPLORE_USERS         = "explore/users"
	EXPLORE_ORGANIZATIONS = "explore/organizations"
	EXPLORE_METADATA      = "explore/metadata"
	DMP_BROWSING          = "explore/dmp_browsing"
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

	// for "Metadata" link on navbar
	c.Data["IsUserFA"] = (c.User.Type >= db.UserFA)

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
func ExploreMetadata(c context.AbstructContext) {
	exploreMetadata(c)
}

// exploreMetadata is RCOS specific code
func exploreMetadata(c context.AbstructContext) {
	c.CallData()["Title"] = c.Tr("explore")
	c.CallData()["PageIsExplore"] = true
	c.CallData()["PageIsExploreMetadata"] = true
	c.CallData()["IsUserFA"] = (c.GetUser().Type >= db.UserFA)

	selectedKey := c.Query("selectKey")
	keyword := c.Query("q")

	page := c.QueryInt("page")
	if page <= 0 {
		page = 1
	}

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

	// Get dmp.json contents
	for _, repo := range repos {
		repo.HasMetadata = false
		gitRepo, repoErr := git.Open(repo.RepoPath())
		if repoErr != nil {
			c.Error(repoErr, "open repository")
			continue
		}

		commit, commintErr := gitRepo.CatFileCommit("refs/heads/master")
		if commintErr != nil || commit == nil {
			log.Error(2, "%s commit could not be retrieved: %v", repo.Name, err)
			c.CallData()["HasDmpJson"] = false
			continue
		}

		entry, err := commit.Blob("/dmp.json")
		if err != nil || entry == nil {
			log.Error(2, "dmp.json blob could not be retrieved: %v", err)
			c.CallData()["HasDmpJson"] = false
			continue
		}
		buf, err := entry.Bytes()
		if err != nil {
			log.Error(2, "dmp.json data could not be read: %v", err)
			c.CallData()["HasDmpJson"] = false
			continue
		}

		c.CallData()["DOIInfo"] = string(buf)

		if selectedKey != "" && keyword != "" && isContained(string(buf), selectedKey, keyword) {
			c.CallData()["SelectedKey"] = selectedKey
			c.CallData()["SearchResult"] = keyword
			repo.HasMetadata = true
		}
	}

	// below is search
	c.CallData()["Keyword"] = keyword
	c.CallData()["Total"] = count
	c.CallData()["Page"] = paginater.New(int(count), conf.UI.ExplorePagingNum, page, 5)

	if err = db.RepositoryList(repos).LoadAttributes(); err != nil {
		c.Error(err, "load attributes")
		return
	}
	c.CallData()["Repos"] = filterUnlistedRepos(repos)

	c.Success(EXPLORE_METADATA)
}

// DmpBrowsing is RCOS specific code
func DmpBrowsing(c context.AbstructContext) {
	page := c.QueryInt("page")
	if page <= 0 {
		page = 1
	}
	repoName := c.Query("repo")
	owner, err := db.Users.GetByUsername(c.Query("owner"))
	if err != nil {
		c.Error(err, "get owner information")
	}

	repos, _, err := db.SearchRepositoryByName(&db.SearchRepoOptions{
		Keyword:  repoName,
		OwnerID:  owner.ID,
		UserID:   c.UserID(),
		OrderBy:  "updated_unix DESC",
		Page:     page,
		PageSize: conf.UI.ExplorePagingNum,
	})
	if err != nil {
		c.Error(err, "search repository by name")
		return
	}

	// Get dmp.json contents
	for _, repo := range repos {
		repo.HasMetadata = false
		gitRepo, repoErr := git.Open(repo.RepoPath())
		if repoErr != nil {
			c.Error(repoErr, "open repository")
			continue
		}

		commit, commintErr := gitRepo.CatFileCommit("refs/heads/master")
		if commintErr != nil || commit == nil {
			log.Error(2, "%s commit could not be retrieved: %v", repo.Name, err)
			c.CallData()["HasDmpJson"] = false
			continue
		}

		entry, err := commit.Blob("/dmp.json")
		if err != nil || entry == nil {
			log.Error(2, "dmp.json blob could not be retrieved: %v", err)
			c.CallData()["HasDmpJson"] = false
			continue
		}
		buf, err := entry.Bytes()
		if err != nil {
			log.Error(2, "dmp.json data could not be read: %v", err)
			c.CallData()["HasDmpJson"] = false
			continue
		}

		c.CallData()["OwnerName"] = owner.Name
		c.CallData()["RepoName"] = repoName
		c.CallData()["DOIInfo"] = string(buf)
	}
	c.Success(DMP_BROWSING)
}

// isContained is RCOS specific code
// ExploreMetadata()で検索結果表示の制御に利用している
func isContained(bufStr, selectedKey, keyword string) bool {
	return strings.Contains(bufStr, "\""+selectedKey+"\": \""+keyword+"\"")
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

	// for "Metadata" link on navbar
	c.Data["IsUserFA"] = (c.User.Type >= db.UserFA)

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

	// for "Metadata" link on navbar
	c.Data["IsUserFA"] = (c.User.Type >= db.UserFA)

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
