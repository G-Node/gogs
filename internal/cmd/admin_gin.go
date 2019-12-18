// Copyright 2016 The Gogs Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package cmd

import (
	"github.com/urfave/cli"

	"github.com/G-Node/gogs/models"
)

var (
	subcmdRebuildSearchIndex = cli.Command{
		Name:  "rebuild-index",
		Usage: "Rebuild the search index for all repositories",
		Action: adminDashboardOperation(
			models.RebuildIndex,
			"Sending all existing repositories to the gin-dex server for reindexing",
		),
		Flags: []cli.Flag{
			stringFlag("config, c", "custom/conf/app.ini", "Custom configuration file path"),
		},
	}
)
