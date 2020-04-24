// Copyright 2016 The Gogs Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package org

import (
	"net/http"

	"github.com/G-Node/gogs/internal/db"
	"github.com/G-Node/gogs/internal/db/errors"
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
	org := c.Org.Organization
	if !org.IsOwnedBy(c.User.ID) {
		c.Error(http.StatusForbidden, "", "given user is not owner of organization")
		return
	}
	team := &db.Team{
		OrgID:       org.ID,
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

func GetTeam(c *context.APIContext) {
	teamname := c.Params(":team")
	org := c.Org.Organization
	team, err := org.GetTeam(teamname)
	if err != nil {
		c.NotFoundOrServerError("GetTeamByName", errors.IsTeamNotExist, err)
		return
	}
	c.JSON(200, convert2.ToTeam(team))
}

func EditTeam(c *context.APIContext, opt api.CreateTeamOption) {
	teamname := c.Params(":team")
	org := c.Org.Organization
	if !org.IsOwnedBy(c.User.ID) {
		c.Error(http.StatusForbidden, "", "given user is not owner of organization")
		return
	}
	oldteam, err := org.GetTeam(teamname)
	if err != nil {
		c.NotFoundOrServerError("EditTeamByName", errors.IsTeamNotExist, err)
		return
	}
	newperm := oldteam.Authorize
	newname := oldteam.Name
	if !oldteam.IsOwnerTeam() {
		// Not allowed to change Owner team perms or name
		newperm = db.ParseAccessMode(opt.Permission)
		newname = opt.Name
	}
	team := &db.Team{
		ID:          oldteam.ID,
		OrgID:       org.ID,
		Name:        newname,
		Description: opt.Description,
		Authorize:   newperm,
	}
	if err := db.UpdateTeam(team, oldteam.Authorize != team.Authorize); err != nil {
		c.NotFoundOrServerError("EditTeamByName", errors.IsTeamNotExist, err)
		return
	}
	c.JSON(http.StatusCreated, convert2.ToTeam(team))
}

func DeleteTeam(c *context.APIContext) {
	teamname := c.Params(":team")
	org := c.Org.Organization
	team, err := org.GetTeam(teamname)
	if err != nil {
		c.NotFoundOrServerError("DeleteTeamByName", errors.IsTeamNotExist, err)
		return
	}
	if team.IsOwnerTeam() {
		c.Error(http.StatusMethodNotAllowed, "", "cannot delete Owners team")
		return
	}
	if err := db.DeleteTeam(team); err != nil {
		c.NotFoundOrServerError("DeleteTeamByName", errors.IsTeamNotExist, err)
		return
	}
	c.NoContent()
}
