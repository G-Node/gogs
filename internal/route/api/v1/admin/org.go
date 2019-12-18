// Copyright 2015 The Gogs Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package admin

import (
	org2 "github.com/G-Node/gogs/internal/route/api/v1/org"
	user2 "github.com/G-Node/gogs/internal/route/api/v1/user"
	api "github.com/gogs/go-gogs-client"

	"github.com/G-Node/gogs/internal/context"
)

func CreateOrg(c *context.APIContext, form api.CreateOrgOption) {
	org2.CreateOrgForUser(c, form, user2.GetUserByParams(c))
}
