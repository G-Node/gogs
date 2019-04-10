// Copyright 2014 The Gogs Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package auth

import (
	"strings"
	"time"

	"github.com/go-macaron/session"
	gouuid "github.com/satori/go.uuid"
	log "gopkg.in/clog.v1"
	"gopkg.in/macaron.v1"

	"github.com/G-Node/gogs/models"
	"github.com/G-Node/gogs/models/errors"
	"github.com/G-Node/gogs/pkg/setting"
	"github.com/G-Node/gogs/pkg/tool"
)

func IsAPIPath(url string) bool {
	return strings.HasPrefix(url, "/api/")
}

// SignedInID returns the id of signed in user, along with one bool value which indicates whether user uses token
// authentication.
func SignedInID(c *macaron.Context, sess session.Store) (_ int64, isTokenAuth bool) {
	if !models.HasEngine {
		return 0, false
	}

	// Check access token.
	if IsAPIPath(c.Req.URL.Path) {
		tokenSHA := c.Query("token")
		if len(tokenSHA) <= 0 {
			tokenSHA = c.Query("access_token")
		}
		if len(tokenSHA) == 0 {
			// Well, check with header again.
			auHead := c.Req.Header.Get("Authorization")
			if len(auHead) > 0 {
				auths := strings.Fields(auHead)
				if len(auths) == 2 && auths[0] == "token" {
					tokenSHA = auths[1]
				}
			}
		}

		// Let's see if token is valid.
		if len(tokenSHA) > 0 {
			t, err := models.GetAccessTokenBySHA(tokenSHA)
			if err != nil {
				if !models.IsErrAccessTokenNotExist(err) && !models.IsErrAccessTokenEmpty(err) {
					log.Error(2, "GetAccessTokenBySHA: %v", err)
				}
				return 0, false
			}
			t.Updated = time.Now()
			if err = models.UpdateAccessToken(t); err != nil {
				log.Error(2, "UpdateAccessToken: %v", err)
			}
			return t.UID, true
		}
	}

	uid := sess.Get("uid")
	if uid == nil {
		return 0, false
	}
	if id, ok := uid.(int64); ok {
		if _, err := models.GetUserByID(id); err != nil {
			if !errors.IsUserNotExist(err) {
				log.Error(2, "GetUserByID: %v", err)
			}
			return 0, false
		}
		return id, false
	}
	return 0, false
}

// SignedInUser returns the user object of signed in user, along with two bool values,
// which indicate whether user uses HTTP Basic Authentication or token authentication respectively.
func SignedInUser(ctx *macaron.Context, sess session.Store) (_ *models.User, isBasicAuth bool, isTokenAuth bool) {
	if !models.HasEngine {
		return nil, false, false
	}

	uid, isTokenAuth := SignedInID(ctx, sess)

	if uid <= 0 {
		if setting.Service.EnableReverseProxyAuth {
			webAuthUser := ctx.Req.Header.Get(setting.ReverseProxyAuthUser)
			if len(webAuthUser) > 0 {
				u, err := models.GetUserByName(webAuthUser)
				if err != nil {
					if !errors.IsUserNotExist(err) {
						log.Error(2, "GetUserByName: %v", err)
						return nil, false, false
					}

					// Check if enabled auto-registration.
					if setting.Service.EnableReverseProxyAutoRegister {
						u := &models.User{
							Name:     webAuthUser,
							Email:    gouuid.NewV4().String() + "@localhost",
							Passwd:   webAuthUser,
							IsActive: true,
						}
						if err = models.CreateUser(u); err != nil {
							// FIXME: should I create a system notice?
							log.Error(2, "CreateUser: %v", err)
							return nil, false, false
						} else {
							return u, false, false
						}
					}
				}
				return u, false, false
			}
		}

		// Check with basic auth.
		baHead := ctx.Req.Header.Get("Authorization")
		if len(baHead) > 0 {
			auths := strings.Fields(baHead)
			if len(auths) == 2 && auths[0] == "Basic" {
				uname, passwd, _ := tool.BasicAuthDecode(auths[1])

				u, err := models.UserLogin(uname, passwd, -1)
				if err != nil {
					if !errors.IsUserNotExist(err) {
						log.Error(2, "UserLogin: %v", err)
					}
					return nil, false, false
				}

				return u, true, false
			}
		}
		return nil, false, false
	}

	u, err := models.GetUserByID(uid)
	if err != nil {
		log.Error(2, "GetUserByID: %v", err)
		return nil, false, false
	}
	return u, false, isTokenAuth
}
