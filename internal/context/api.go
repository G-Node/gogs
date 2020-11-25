// Copyright 2016 The Gogs Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package context

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/unknwon/paginater"
	"gopkg.in/macaron.v1"
	log "unknwon.dev/clog/v2"

	"github.com/G-Node/gogs/internal/conf"
)

type APIContext struct {
	*Context // TODO: Reduce to only needed fields instead of full shadow

	// Base URL for the version of API endpoints, e.g. https://try.github.com/G-Node/api/v1
	BaseURL string

	Org *APIOrganization
}

// FIXME: move this constant to github.com/gogs/go-gogs-client
const DocURL = "https://github.com/gogs/docs-api"

// Error responses error message to client with given message.
// If status is 500, also it prints error to log.
func (c *APIContext) Error(status int, title string, obj interface{}) {
	var message string
	if err, ok := obj.(error); ok {
		message = err.Error()
	} else {
		message = obj.(string)
	}

	if status == http.StatusInternalServerError {
		log.ErrorDepth(5, "%s: %s", title, message)
	}

	c.JSON(status, map[string]string{
		"message": message,
		"url":     DocURL,
	})
}

// NoContent renders the 204 response.
func (c *APIContext) NoContent() {
	c.Status(http.StatusNoContent)
}

// NotFound renders the 404 response.
func (c *APIContext) NotFound() {
	c.Status(http.StatusNotFound)
}

// ServerError renders the 500 response.
func (c *APIContext) ServerError(title string, err error) {
	c.Error(http.StatusInternalServerError, title, err)
}

// Errorf renders the 500 response with formatted message.
func (c *APIContext) Errorf(err error, format string, args ...interface{}) {
	c.Error(http.StatusInternalServerError, fmt.Sprintf(format, args...), err)
}

// NotFoundOrServerError use error check function to determine if the error
// is about not found. It responses with 404 status code for not found error,
// or error context description for logging purpose of 500 server error.
func (c *APIContext) NotFoundOrServerError(title string, errck func(error) bool, err error) {
	if errck(err) {
		c.NotFound()
		return
	}
	c.ServerError(title, err)
}

// SetLinkHeader sets pagination link header by given total number and page size.
func (c *APIContext) SetLinkHeader(total, pageSize int) {
	page := paginater.New(total, pageSize, c.QueryInt("page"), 0)
	links := make([]string, 0, 4)
	if page.HasNext() {
		links = append(links, fmt.Sprintf("<%s%s?page=%d>; rel=\"next\"", conf.Server.ExternalURL, c.Req.URL.Path[1:], page.Next()))
	}
	if !page.IsLast() {
		links = append(links, fmt.Sprintf("<%s%s?page=%d>; rel=\"last\"", conf.Server.ExternalURL, c.Req.URL.Path[1:], page.TotalPages()))
	}
	if !page.IsFirst() {
		links = append(links, fmt.Sprintf("<%s%s?page=1>; rel=\"first\"", conf.Server.ExternalURL, c.Req.URL.Path[1:]))
	}
	if page.HasPrevious() {
		links = append(links, fmt.Sprintf("<%s%s?page=%d>; rel=\"prev\"", conf.Server.ExternalURL, c.Req.URL.Path[1:], page.Previous()))
	}

	if len(links) > 0 {
		c.Header().Set("Link", strings.Join(links, ","))
	}
}

func APIContexter() macaron.Handler {
	return func(ctx *Context) {
		c := &APIContext{
			Context: ctx,
			BaseURL: conf.Server.ExternalURL + "api/v1",
		}
		ctx.Map(c)
	}
}
