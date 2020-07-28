// +build go1.11

// Copyright 2014 The Gogs Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

// Gogs is a painless self-hosted Git Service.
package main

import (
	"os"

	"github.com/urfave/cli"
	log "gopkg.in/clog.v1"

	"github.com/G-Node/gogs/internal/cmd"
	"github.com/G-Node/gogs/internal/setting"
)

const Version = "0.12.0+dev"

func init() {
	setting.AppVer = Version
}

func main() {
	app := cli.NewApp()
	app.Name = "GIN"
	app.Usage = "Modern Research Data Management for Neuroscience"
	app.Version = Version
	app.Commands = []cli.Command{
		cmd.Web,
		cmd.Serv,
		cmd.Hook,
		cmd.Cert,
		cmd.Admin,
		cmd.Import,
		cmd.Backup,
		cmd.Restore,
	}
	if err := app.Run(os.Args); err != nil {
		log.Fatal(2, "Failed to run: %v", err)
	}
}
