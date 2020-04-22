// Copyright 2016 The Gogs Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package org

import (
	"net/http"

	"github.com/G-Node/gogs/internal/db"
	convert2 "github.com/G-Node/gogs/internal/route/api/v1/convert"
	api "github.com/gogs/go-gogs-client"

	"github.com/G-Node/gogs/internal/context"
)

func ListTeams(c *context.APIContext) {
	org := c.Org.Organization
	if err := org.GetTeams(); err != nil {
		c.Error(500, "GetTeams", err)
		return
	}

	apiTeams := make([]*api.Team, len(org.Teams))
	for i := range org.Teams {
		apiTeams[i] = convert2.ToTeam(org.Teams[i])
	}
	c.JSON(200, apiTeams)
}

func CreateTeam(c *context.APIContext, opt api.CreateTeamOption) {
	if !c.Org.Organization.IsOwnedBy(c.User.ID) {
		c.Error(http.StatusForbidden, "", "given user is not owner of organization")
		return
	}
	team := &db.Team{
		OrgID:       c.Org.Organization.ID,
		Name:        opt.Name,
		Description: opt.Description,
		Authorize:   db.ParseAccessMode(opt.Permission),
	}
	if err := db.NewTeam(team); err != nil {
		if db.IsErrTeamAlreadyExist(err) {
			c.Error(http.StatusUnprocessableEntity, "", err)
		} else {
			c.ServerError("NewTeam", err)
		}
		return
	}

	c.JSON(http.StatusCreated, convert2.ToTeam(team))
}
