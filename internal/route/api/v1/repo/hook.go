// Copyright 2014 The Gogs Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package repo

import (
	convert2 "github.com/G-Node/gogs/internal/route/api/v1/convert"
	"github.com/json-iterator/go"
	"github.com/unknwon/com"

	api "github.com/gogs/go-gogs-client"

	"github.com/G-Node/gogs/internal/context"
	"github.com/G-Node/gogs/internal/db"
	"github.com/G-Node/gogs/internal/db/errors"
)

// https://github.com/gogs/go-gogs-client/wiki/Repositories#list-hooks
func ListHooks(c *context.APIContext) {
	hooks, err := db.GetWebhooksByRepoID(c.Repo.Repository.ID)
	if err != nil {
		c.Error(500, "GetWebhooksByRepoID", err)
		return
	}

	apiHooks := make([]*api.Hook, len(hooks))
	for i := range hooks {
		apiHooks[i] = convert2.ToHook(c.Repo.RepoLink, hooks[i])
	}
	c.JSON(200, &apiHooks)
}

// https://github.com/gogs/go-gogs-client/wiki/Repositories#create-a-hook
func CreateHook(c *context.APIContext, form api.CreateHookOption) {
	if !db.IsValidHookTaskType(form.Type) {
		c.Error(422, "", "Invalid hook type")
		return
	}
	for _, name := range []string{"url", "content_type"} {
		if _, ok := form.Config[name]; !ok {
			c.Error(422, "", "Missing config option: "+name)
			return
		}
	}
	if !db.IsValidHookContentType(form.Config["content_type"]) {
		c.Error(422, "", "Invalid content type")
		return
	}

	if len(form.Events) == 0 {
		form.Events = []string{"push"}
	}
	w := &db.Webhook{
		RepoID:      c.Repo.Repository.ID,
		URL:         form.Config["url"],
		ContentType: db.ToHookContentType(form.Config["content_type"]),
		Secret:      form.Config["secret"],
		HookEvent: &db.HookEvent{
			ChooseEvents: true,
			HookEvents: db.HookEvents{
				Create:       com.IsSliceContainsStr(form.Events, string(db.HOOK_EVENT_CREATE)),
				Delete:       com.IsSliceContainsStr(form.Events, string(db.HOOK_EVENT_DELETE)),
				Fork:         com.IsSliceContainsStr(form.Events, string(db.HOOK_EVENT_FORK)),
				Push:         com.IsSliceContainsStr(form.Events, string(db.HOOK_EVENT_PUSH)),
				Issues:       com.IsSliceContainsStr(form.Events, string(db.HOOK_EVENT_ISSUES)),
				IssueComment: com.IsSliceContainsStr(form.Events, string(db.HOOK_EVENT_ISSUE_COMMENT)),
				PullRequest:  com.IsSliceContainsStr(form.Events, string(db.HOOK_EVENT_PULL_REQUEST)),
				Release:      com.IsSliceContainsStr(form.Events, string(db.HOOK_EVENT_RELEASE)),
			},
		},
		IsActive:     form.Active,
		HookTaskType: db.ToHookTaskType(form.Type),
	}
	if w.HookTaskType == db.SLACK {
		channel, ok := form.Config["channel"]
		if !ok {
			c.Error(422, "", "Missing config option: channel")
			return
		}
		meta, err := jsoniter.Marshal(&db.SlackMeta{
			Channel:  channel,
			Username: form.Config["username"],
			IconURL:  form.Config["icon_url"],
			Color:    form.Config["color"],
		})
		if err != nil {
			c.Error(500, "slack: JSON marshal failed", err)
			return
		}
		w.Meta = string(meta)
	}

	if err := w.UpdateEvent(); err != nil {
		c.Error(500, "UpdateEvent", err)
		return
	} else if err := db.CreateWebhook(w); err != nil {
		c.Error(500, "CreateWebhook", err)
		return
	}

	c.JSON(201, convert2.ToHook(c.Repo.RepoLink, w))
}

// https://github.com/gogs/go-gogs-client/wiki/Repositories#edit-a-hook
func EditHook(c *context.APIContext, form api.EditHookOption) {
	w, err := db.GetWebhookOfRepoByID(c.Repo.Repository.ID, c.ParamsInt64(":id"))
	if err != nil {
		if errors.IsWebhookNotExist(err) {
			c.Status(404)
		} else {
			c.Error(500, "GetWebhookOfRepoByID", err)
		}
		return
	}

	if form.Config != nil {
		if url, ok := form.Config["url"]; ok {
			w.URL = url
		}
		if ct, ok := form.Config["content_type"]; ok {
			if !db.IsValidHookContentType(ct) {
				c.Error(422, "", "Invalid content type")
				return
			}
			w.ContentType = db.ToHookContentType(ct)
		}

		if w.HookTaskType == db.SLACK {
			if channel, ok := form.Config["channel"]; ok {
				meta, err := jsoniter.Marshal(&db.SlackMeta{
					Channel:  channel,
					Username: form.Config["username"],
					IconURL:  form.Config["icon_url"],
					Color:    form.Config["color"],
				})
				if err != nil {
					c.Error(500, "slack: JSON marshal failed", err)
					return
				}
				w.Meta = string(meta)
			}
		}
	}

	// Update events
	if len(form.Events) == 0 {
		form.Events = []string{"push"}
	}
	w.PushOnly = false
	w.SendEverything = false
	w.ChooseEvents = true
	w.Create = com.IsSliceContainsStr(form.Events, string(db.HOOK_EVENT_CREATE))
	w.Delete = com.IsSliceContainsStr(form.Events, string(db.HOOK_EVENT_DELETE))
	w.Fork = com.IsSliceContainsStr(form.Events, string(db.HOOK_EVENT_FORK))
	w.Push = com.IsSliceContainsStr(form.Events, string(db.HOOK_EVENT_PUSH))
	w.Issues = com.IsSliceContainsStr(form.Events, string(db.HOOK_EVENT_ISSUES))
	w.IssueComment = com.IsSliceContainsStr(form.Events, string(db.HOOK_EVENT_ISSUE_COMMENT))
	w.PullRequest = com.IsSliceContainsStr(form.Events, string(db.HOOK_EVENT_PULL_REQUEST))
	w.Release = com.IsSliceContainsStr(form.Events, string(db.HOOK_EVENT_RELEASE))
	if err = w.UpdateEvent(); err != nil {
		c.Error(500, "UpdateEvent", err)
		return
	}

	if form.Active != nil {
		w.IsActive = *form.Active
	}

	if err := db.UpdateWebhook(w); err != nil {
		c.Error(500, "UpdateWebhook", err)
		return
	}

	c.JSON(200, convert2.ToHook(c.Repo.RepoLink, w))
}

func DeleteHook(c *context.APIContext) {
	if err := db.DeleteWebhookOfRepoByID(c.Repo.Repository.ID, c.ParamsInt64(":id")); err != nil {
		c.Error(500, "DeleteWebhookByRepoID", err)
		return
	}

	c.Status(204)
}
