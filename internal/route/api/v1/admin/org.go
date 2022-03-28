// Copyright 2015 The Gogs Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package admin

import (
	api "github.com/gogs/go-gogs-client"

	"github.com/NII-DG/gogs/internal/context"
	"github.com/NII-DG/gogs/internal/route/api/v1/org"
	"github.com/NII-DG/gogs/internal/route/api/v1/user"
)

func CreateOrg(c *context.APIContext, form api.CreateOrgOption) {
	org.CreateOrgForUser(c, form, user.GetUserByParams(c))
}
