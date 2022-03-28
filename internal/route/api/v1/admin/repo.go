// Copyright 2015 The Gogs Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package admin

import (
	api "github.com/gogs/go-gogs-client"

	"github.com/NII-DG/gogs/internal/context"
	"github.com/NII-DG/gogs/internal/route/api/v1/repo"
	"github.com/NII-DG/gogs/internal/route/api/v1/user"
)

func CreateRepo(c *context.APIContext, form api.CreateRepoOption) {
	owner := user.GetUserByParams(c)
	if c.Written() {
		return
	}

	repo.CreateUserRepo(c, owner, form)
}
