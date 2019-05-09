package dav

import (
	"net/http"
	"strings"

	"github.com/G-Node/git-module"
	"github.com/G-Node/gogs/models"
	"github.com/G-Node/gogs/pkg/context"
	"gopkg.in/macaron.v1"
)

// DavMiddle initialises and returns a WebDav middleware handler (macaron.Handler)
// [0]: issues, [1]: wiki
func DavMiddle() macaron.Handler {
	return func(c *context.Context) {
		var (
			owner *models.User
			err   error
		)

		ownerName := c.Params(":username")
		repoName := strings.TrimSuffix(c.Params(":reponame"), ".git")

		// Check if the user is the same as the repository owner
		if c.IsLogged && c.User.LowerName == strings.ToLower(ownerName) {
			owner = c.User
		} else {
			owner, err = models.GetUserByName(ownerName)
			if err != nil {
				Webdav401(c)
				return
			}
		}
		c.Repo.Owner = owner

		repo, err := models.GetRepositoryByName(owner.ID, repoName)
		if err != nil {
			Webdav401(c)
			return
		}

		c.Repo.Repository = repo
		c.Repo.RepoLink = repo.Link()

		// Admin has super access.
		if c.IsLogged && c.User.IsAdmin {
			c.Repo.AccessMode = models.ACCESS_MODE_OWNER
		} else {
			mode, err := models.AccessLevel(c.UserID(), repo)
			if err != nil {
				c.WriteHeader(http.StatusInternalServerError)
				return
			}
			c.Repo.AccessMode = mode
		}

		if repo.IsMirror {
			c.Repo.Mirror, err = models.GetMirrorByRepoID(repo.ID)
			if err != nil {
				c.WriteHeader(http.StatusInternalServerError)
				return
			}
		}

		gitRepo, err := git.OpenRepository(models.RepoPath(ownerName, repoName))
		if err != nil {
			c.WriteHeader(http.StatusInternalServerError)
			return
		}
		c.Repo.GitRepo = gitRepo

		tags, err := c.Repo.GitRepo.GetTags()
		if err != nil {
			c.WriteHeader(http.StatusInternalServerError)
			return
		}
		c.Repo.Repository.NumTags = len(tags)

		// repo is bare and display enable
		if c.Repo.Repository.IsBare {
			return
		}

		brs, err := c.Repo.GitRepo.GetBranches()
		if err != nil {
			c.WriteHeader(http.StatusInternalServerError)
			return
		}
		// If not branch selected, try default one.
		// If default branch doesn't exists, fall back to some other branch.
		if len(c.Repo.BranchName) == 0 {
			if len(c.Repo.Repository.DefaultBranch) > 0 && gitRepo.IsBranchExist(c.Repo.Repository.DefaultBranch) {
				c.Repo.BranchName = c.Repo.Repository.DefaultBranch
			} else if len(brs) > 0 {
				c.Repo.BranchName = brs[0]
			}
		}
	}
}
