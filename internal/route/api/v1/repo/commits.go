// Copyright 2018 The Gogs Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package repo

import (
	"net/http"
	"strings"
	"time"

	"github.com/gogs/git-module"
	api "github.com/gogs/go-gogs-client"

	"github.com/NII-DG/gogs/internal/conf"
	"github.com/NII-DG/gogs/internal/context"
	"github.com/NII-DG/gogs/internal/db"
	"github.com/NII-DG/gogs/internal/gitutil"
)

func GetSingleCommit(c *context.APIContext) {
	if strings.Contains(c.Req.Header.Get("Accept"), api.MediaApplicationSHA) {
		c.SetParams("*", c.Params(":sha"))
		GetReferenceSHA(c)
		return
	}

	gitRepo, err := git.Open(c.Repo.Repository.RepoPath())
	if err != nil {
		c.Error(err, "open repository")
		return
	}
	commit, err := gitRepo.CatFileCommit(c.Params(":sha"))
	if err != nil {
		c.NotFoundOrError(gitutil.NewError(err), "get commit")
		return
	}

	// Retrieve author and committer information
	var apiAuthor, apiCommitter *api.User
	author, err := db.GetUserByEmail(commit.Author.Email)
	if err != nil && !db.IsErrUserNotExist(err) {
		c.Error(err, "get user by author email")
		return
	} else if err == nil {
		apiAuthor = author.APIFormat()
	}
	// Save one query if the author is also the committer
	if commit.Committer.Email == commit.Author.Email {
		apiCommitter = apiAuthor
	} else {
		committer, err := db.GetUserByEmail(commit.Committer.Email)
		if err != nil && !db.IsErrUserNotExist(err) {
			c.Error(err, "get user by committer email")
			return
		} else if err == nil {
			apiCommitter = committer.APIFormat()
		}
	}

	// Retrieve parent(s) of the commit
	apiParents := make([]*api.CommitMeta, commit.ParentsCount())
	for i := 0; i < commit.ParentsCount(); i++ {
		sha, _ := commit.ParentID(i)
		apiParents[i] = &api.CommitMeta{
			URL: c.BaseURL + "/repos/" + c.Repo.Repository.FullName() + "/commits/" + sha.String(),
			SHA: sha.String(),
		}
	}

	c.JSONSuccess(&api.Commit{
		CommitMeta: &api.CommitMeta{
			URL: conf.Server.ExternalURL + c.Link[1:],
			SHA: commit.ID.String(),
		},
		HTMLURL: c.Repo.Repository.HTMLURL() + "/commits/" + commit.ID.String(),
		RepoCommit: &api.RepoCommit{
			URL: conf.Server.ExternalURL + c.Link[1:],
			Author: &api.CommitUser{
				Name:  commit.Author.Name,
				Email: commit.Author.Email,
				Date:  commit.Author.When.Format(time.RFC3339),
			},
			Committer: &api.CommitUser{
				Name:  commit.Committer.Name,
				Email: commit.Committer.Email,
				Date:  commit.Committer.When.Format(time.RFC3339),
			},
			Message: commit.Summary(),
			Tree: &api.CommitMeta{
				URL: c.BaseURL + "/repos/" + c.Repo.Repository.FullName() + "/tree/" + commit.ID.String(),
				SHA: commit.ID.String(),
			},
		},
		Author:    apiAuthor,
		Committer: apiCommitter,
		Parents:   apiParents,
	})
}

func GetReferenceSHA(c *context.APIContext) {
	gitRepo, err := git.Open(c.Repo.Repository.RepoPath())
	if err != nil {
		c.Error(err, "open repository")
		return
	}

	ref := c.Params("*")
	refType := 0 // 0-unknown, 1-branch, 2-tag
	if strings.HasPrefix(ref, git.RefsHeads) {
		ref = strings.TrimPrefix(ref, git.RefsHeads)
		refType = 1
	} else if strings.HasPrefix(ref, git.RefsTags) {
		ref = strings.TrimPrefix(ref, git.RefsTags)
		refType = 2
	} else {
		if gitRepo.HasBranch(ref) {
			refType = 1
		} else if gitRepo.HasTag(ref) {
			refType = 2
		} else {
			c.NotFound()
			return
		}
	}

	var sha string
	if refType == 1 {
		sha, err = gitRepo.BranchCommitID(ref)
	} else if refType == 2 {
		sha, err = gitRepo.TagCommitID(ref)
	}
	if err != nil {
		c.NotFoundOrError(gitutil.NewError(err), "get reference commit ID")
		return
	}
	c.PlainText(http.StatusOK, sha)
}
