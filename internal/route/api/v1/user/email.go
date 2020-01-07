// Copyright 2015 The Gogs Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package user

import (
	convert2 "github.com/G-Node/gogs/internal/route/api/v1/convert"
	"net/http"

	api "github.com/gogs/go-gogs-client"

	"github.com/G-Node/gogs/internal/context"
	"github.com/G-Node/gogs/internal/db"
	"github.com/G-Node/gogs/internal/setting"
)

func ListEmails(c *context.APIContext) {
	emails, err := db.GetEmailAddresses(c.User.ID)
	if err != nil {
		c.ServerError("GetEmailAddresses", err)
		return
	}
	apiEmails := make([]*api.Email, len(emails))
	for i := range emails {
		apiEmails[i] = convert2.ToEmail(emails[i])
	}
	c.JSONSuccess(&apiEmails)
}

func AddEmail(c *context.APIContext, form api.CreateEmailOption) {
	if len(form.Emails) == 0 {
		c.Status(http.StatusUnprocessableEntity)
		return
	}

	emails := make([]*db.EmailAddress, len(form.Emails))
	for i := range form.Emails {
		emails[i] = &db.EmailAddress{
			UID:         c.User.ID,
			Email:       form.Emails[i],
			IsActivated: !setting.Service.RegisterEmailConfirm,
		}
	}

	if err := db.AddEmailAddresses(emails); err != nil {
		if db.IsErrEmailAlreadyUsed(err) {
			c.Error(http.StatusUnprocessableEntity, "", "email address has been used: "+err.(db.ErrEmailAlreadyUsed).Email)
		} else {
			c.Error(http.StatusInternalServerError, "AddEmailAddresses", err)
		}
		return
	}

	apiEmails := make([]*api.Email, len(emails))
	for i := range emails {
		apiEmails[i] = convert2.ToEmail(emails[i])
	}
	c.JSON(http.StatusCreated, &apiEmails)
}

func DeleteEmail(c *context.APIContext, form api.CreateEmailOption) {
	if len(form.Emails) == 0 {
		c.NoContent()
		return
	}

	emails := make([]*db.EmailAddress, len(form.Emails))
	for i := range form.Emails {
		emails[i] = &db.EmailAddress{
			UID:   c.User.ID,
			Email: form.Emails[i],
		}
	}

	if err := db.DeleteEmailAddresses(emails); err != nil {
		c.Error(http.StatusInternalServerError, "DeleteEmailAddresses", err)
		return
	}
	c.NoContent()
}
