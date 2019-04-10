// Copyright 2016 The Gogs Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package context

import (
	"github.com/G-Node/gogs/models"
)

type APIOrganization struct {
	Organization *models.User
	Team         *models.Team
}
