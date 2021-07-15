// Copyright 2017 The Gogs Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package migrations

import (
	"fmt"
	"path/filepath"
	"strings"

	log "unknwon.dev/clog/v2"
	"xorm.io/xorm"

	"github.com/gogs/git-module"

	"github.com/ivis-yoshida/gogs/internal/conf"
)

func updateRepositorySizes(x *xorm.Engine) (err error) {
	log.Info("[migrations.v16] This migration could take up to minutes, please be patient.")
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
		if err = x.SQL(fmt.Sprintf("SELECT * FROM `repository` ORDER BY id ASC LIMIT 10 OFFSET %d", offset)).
			Find(&repos); err != nil {
			return fmt.Errorf("select repos [offset: %d]: %v", offset, err)
		}
		log.Trace("[migrations.v16] Select [offset: %d, repos: %d]", offset, len(repos))
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

			repoPath := strings.ToLower(filepath.Join(conf.Repository.Root, user.Name, repo.Name)) + ".git"
			countObject, err := git.RepoCountObjects(repoPath)
			if err != nil {
				log.Warn("[migrations.v16] Count repository objects: %v", err)
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
