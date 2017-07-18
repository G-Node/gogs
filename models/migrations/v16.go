// Copyright 2017 The Gogs Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package migrations

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/go-xorm/xorm"
	log "gopkg.in/clog.v1"

	"github.com/gogits/git-module"

	"github.com/G-Node/gogs/pkg/setting"
)

func updateRepositorySizes(x *xorm.Engine) (err error) {
	log.Info("This migration could take up to minutes, please be patient.")
	type Repository struct {
		ID      int64
		OwnerID int64
		Name    string
		Size    int64
	}
	type User struct {
		ID   int64
		Name string
	}
	if err = x.Sync2(new(Repository)); err != nil {
		return fmt.Errorf("Sync2: %v", err)
	}

	// For the sake of SQLite3, we can't use x.Iterate here.
	offset := 0
	for {
		repos := make([]*Repository, 0, 10)
		if err = x.Sql(fmt.Sprintf("SELECT * FROM `repository` ORDER BY id ASC LIMIT 10 OFFSET %d", offset)).
			Find(&repos); err != nil {
			return fmt.Errorf("select repos [offset: %d]: %v", offset, err)
		}
		log.Trace("Select [offset: %d, repos: %d]", offset, len(repos))
		if len(repos) == 0 {
			break
		}
		offset += 10

		for _, repo := range repos {
			if repo.Name == "." || repo.Name == ".." {
				continue
			}

			user := new(User)
			has, err := x.Where("id = ?", repo.OwnerID).Get(user)
			if err != nil {
				return fmt.Errorf("query owner of repository [repo_id: %d, owner_id: %d]: %v", repo.ID, repo.OwnerID, err)
			} else if !has {
				continue
			}

			repoPath := filepath.Join(setting.RepoRootPath, strings.ToLower(user.Name), strings.ToLower(repo.Name)) + ".git"
			countObject, err := git.GetRepoSize(repoPath)
			if err != nil {
				log.Warn("GetRepoSize: %v", err)
				continue
			}

			repo.Size = countObject.Size + countObject.SizePack
			if _, err = x.Id(repo.ID).Cols("size").Update(repo); err != nil {
				return fmt.Errorf("update size: %v", err)
			}
		}
	}
	return nil
}
