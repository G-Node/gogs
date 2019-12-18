// Copyright 2015 The Gogs Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package admin

import (
	repo2 "github.com/G-Node/gogs/internal/route/api/v1/repo"
	user2 "github.com/G-Node/gogs/internal/route/api/v1/user"
	api "github.com/gogs/go-gogs-client"

	"github.com/G-Node/gogs/internal/context"
)

func CreateRepo(c *context.APIContext, form api.CreateRepoOption) {
	owner := user2.GetUserByParams(c)
	if c.Written() {
		return
	}

	repo2.CreateUserRepo(c, owner, form)
}
