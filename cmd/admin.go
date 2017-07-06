// Copyright 2016 The Gogs Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package cmd

import (
	"fmt"

	"github.com/urfave/cli"

	"github.com/G-Node/gogs/models"
	"github.com/G-Node/gogs/pkg/setting"
)

var (
	Admin = cli.Command{
		Name:  "admin",
		Usage: "Perform admin operations on command line",
		Description: `Allow using internal logic of Gogs without hacking into the source code
to make automatic initialization process more smoothly`,
		Subcommands: []cli.Command{
			subcmdCreateUser,
		},
	}

	subcmdCreateUser = cli.Command{
		Name:   "create-user",
		Usage:  "Create a new user in database",
		Action: runCreateUser,
		Flags: []cli.Flag{
			stringFlag("name", "", "Username"),
			stringFlag("password", "", "User password"),
			stringFlag("email", "", "User email address"),
			boolFlag("admin", "User is an admin"),
			stringFlag("config, c", "custom/conf/app.ini", "Custom configuration file path"),
		},
	}
)

func runCreateUser(c *cli.Context) error {
	if !c.IsSet("name") {
		return fmt.Errorf("Username is not specified")
	} else if !c.IsSet("password") {
		return fmt.Errorf("Password is not specified")
	} else if !c.IsSet("email") {
		return fmt.Errorf("Email is not specified")
	}

	if c.IsSet("config") {
		setting.CustomConf = c.String("config")
	}

	setting.NewContext()
	models.LoadConfigs()
	models.SetEngine()

	if err := models.CreateUser(&models.User{
		Name:     c.String("name"),
		Email:    c.String("email"),
		Passwd:   c.String("password"),
		IsActive: true,
		IsAdmin:  c.Bool("admin"),
	}); err != nil {
		return fmt.Errorf("CreateUser: %v", err)
	}

	fmt.Printf("New user '%s' has been successfully created!\n", c.String("name"))
	return nil
}
