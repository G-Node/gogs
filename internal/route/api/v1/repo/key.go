// Copyright 2015 The Gogs Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package repo

import (
	"fmt"
	convert2 "github.com/G-Node/gogs/internal/route/api/v1/convert"

	api "github.com/gogs/go-gogs-client"

	"github.com/G-Node/gogs/internal/conf"
	"github.com/G-Node/gogs/internal/context"
	"github.com/G-Node/gogs/internal/db"
)

func composeDeployKeysAPILink(repoPath string) string {
	return conf.Server.ExternalURL + "api/v1/repos/" + repoPath + "/keys/"
}

// https://github.com/gogs/go-gogs-client/wiki/Repositories-Deploy-Keys#list-deploy-keys
func ListDeployKeys(c *context.APIContext) {
	keys, err := db.ListDeployKeys(c.Repo.Repository.ID)
	if err != nil {
		c.Error(500, "ListDeployKeys", err)
		return
	}

	apiLink := composeDeployKeysAPILink(c.Repo.Owner.Name + "/" + c.Repo.Repository.Name)
	apiKeys := make([]*api.DeployKey, len(keys))
	for i := range keys {
		if err = keys[i].GetContent(); err != nil {
			c.Error(500, "GetContent", err)
			return
		}
		apiKeys[i] = convert2.ToDeployKey(apiLink, keys[i])
	}

	c.JSON(200, &apiKeys)
}

// https://github.com/gogs/go-gogs-client/wiki/Repositories-Deploy-Keys#get-a-deploy-key
func GetDeployKey(c *context.APIContext) {
	key, err := db.GetDeployKeyByID(c.ParamsInt64(":id"))
	if err != nil {
		if db.IsErrDeployKeyNotExist(err) {
			c.Status(404)
		} else {
			c.Error(500, "GetDeployKeyByID", err)
		}
		return
	}

	if err = key.GetContent(); err != nil {
		c.Error(500, "GetContent", err)
		return
	}

	apiLink := composeDeployKeysAPILink(c.Repo.Owner.Name + "/" + c.Repo.Repository.Name)
	c.JSON(200, convert2.ToDeployKey(apiLink, key))
}

func HandleCheckKeyStringError(c *context.APIContext, err error) {
	if db.IsErrKeyUnableVerify(err) {
		c.Error(422, "", "Unable to verify key content")
	} else {
		c.Error(422, "", fmt.Errorf("Invalid key content: %v", err))
	}
}

func HandleAddKeyError(c *context.APIContext, err error) {
	switch {
	case db.IsErrKeyAlreadyExist(err):
		c.Error(422, "", "Key content has been used as non-deploy key")
	case db.IsErrKeyNameAlreadyUsed(err):
		c.Error(422, "", "Key title has been used")
	default:
		c.Error(500, "AddKey", err)
	}
}

// https://github.com/gogs/go-gogs-client/wiki/Repositories-Deploy-Keys#add-a-new-deploy-key
func CreateDeployKey(c *context.APIContext, form api.CreateKeyOption) {
	content, err := db.CheckPublicKeyString(form.Key)
	if err != nil {
		HandleCheckKeyStringError(c, err)
		return
	}

	key, err := db.AddDeployKey(c.Repo.Repository.ID, form.Title, content)
	if err != nil {
		HandleAddKeyError(c, err)
		return
	}

	key.Content = content
	apiLink := composeDeployKeysAPILink(c.Repo.Owner.Name + "/" + c.Repo.Repository.Name)
	c.JSON(201, convert2.ToDeployKey(apiLink, key))
}

// https://github.com/gogs/go-gogs-client/wiki/Repositories-Deploy-Keys#remove-a-deploy-key
func DeleteDeploykey(c *context.APIContext) {
	if err := db.DeleteDeployKey(c.User, c.ParamsInt64(":id")); err != nil {
		if db.IsErrKeyAccessDenied(err) {
			c.Error(403, "", "You do not have access to this key")
		} else {
			c.Error(500, "DeleteDeployKey", err)
		}
		return
	}

	c.Status(204)
}
