// Copyright 2016 The Gogs Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package org

import (
	"net/http"

	api "github.com/gogs/go-gogs-client"

	"github.com/ivis-yoshida/gogs/internal/context"
	"github.com/ivis-yoshida/gogs/internal/db"
	"github.com/ivis-yoshida/gogs/internal/db/errors"
	"github.com/ivis-yoshida/gogs/internal/route/api/v1/convert"
)

func ListTeams(c *context.APIContext) {
	org := c.Org.Organization
	if err := org.GetTeams(); err != nil {
		c.Error(err, "get teams")
		return
	}

	apiTeams := make([]*api.Team, len(org.Teams))
	for i := range org.Teams {
		apiTeams[i] = convert.ToTeam(org.Teams[i])
	}
	c.JSONSuccess(apiTeams)
}

func CreateTeam(c *context.APIContext, opt api.CreateTeamOption) {
	org := c.Org.Organization
	if !org.IsOwnedBy(c.User.ID) {
		c.ErrorStatus(http.StatusForbidden, errors.New("given user is not owner of organization"))
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
			c.ErrorStatus(http.StatusUnprocessableEntity, err)
		} else {
			c.Error(err, "NewTeam")
		}
		return
	}

	c.JSON(http.StatusCreated, convert.ToTeam(team))
}

func GetTeam(c *context.APIContext) {
	teamname := c.Params(":team")
	org := c.Org.Organization
	team, err := org.GetTeam(teamname)
	if err != nil {
		c.NotFoundOrError(err, "GetTeamByName")
		return
	}
	c.JSON(200, convert.ToTeam(team))
}

func EditTeam(c *context.APIContext, opt api.CreateTeamOption) {
	teamname := c.Params(":team")
	org := c.Org.Organization
	if !org.IsOwnedBy(c.User.ID) {
		c.ErrorStatus(http.StatusForbidden, errors.New("given user is not owner of organization"))
		return
	}
	oldteam, err := org.GetTeam(teamname)
	if err != nil {
		c.NotFoundOrError(err, "EditTeamByName")
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
		c.NotFoundOrError(err, "EditTeamByName")
		return
	}
	c.JSON(http.StatusCreated, convert.ToTeam(team))
}

func DeleteTeam(c *context.APIContext) {
	teamname := c.Params(":team")
	org := c.Org.Organization
	team, err := org.GetTeam(teamname)
	if err != nil {
		c.NotFoundOrError(err, "DeleteTeamByName")
		return
	}
	if team.IsOwnerTeam() {
		c.ErrorStatus(http.StatusMethodNotAllowed, errors.New("cannot delete Owners team"))
		return
	}
	if err := db.DeleteTeam(team); err != nil {
		c.NotFoundOrError(err, "DeleteTeamByName")
		return
	}
	c.NoContent()
}
