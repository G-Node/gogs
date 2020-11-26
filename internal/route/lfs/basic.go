// Copyright 2020 The Gogs Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package lfs

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"strconv"

	log "unknwon.dev/clog/v2"

	"github.com/G-Node/gogs/internal/conf"
	"github.com/G-Node/gogs/internal/context"
	"github.com/G-Node/gogs/internal/db"
	"github.com/G-Node/gogs/internal/lfsutil"
	"github.com/G-Node/gogs/internal/strutil"
)

const transferBasic = "basic"
const (
	basicOperationUpload   = "upload"
	basicOperationDownload = "download"
)

// GET /{owner}/{repo}.git/info/lfs/object/basic/{oid}
func serveBasicDownload(c *context.Context, repo *db.Repository, oid lfsutil.OID) {
	object, err := db.LFS.GetObjectByOID(repo.ID, oid)
	if err != nil {
		if db.IsErrLFSObjectNotExist(err) {
			responseJSON(c.Resp, http.StatusNotFound, responseError{
				Message: "Object does not exist",
			})
		} else {
			internalServerError(c.Resp)
			log.Error("Failed to get object [repo_id: %d, oid: %s]: %v", repo.ID, oid, err)
		}
		return
	}

	fpath := lfsutil.StorageLocalPath(conf.LFS.ObjectsPath, object.OID)
	r, err := os.Open(fpath)
	if err != nil {
		internalServerError(c.Resp)
		log.Error("Failed to open object file [path: %s]: %v", fpath, err)
		return
	}
	defer r.Close()

	c.Header().Set("Content-Type", "application/octet-stream")
	c.Header().Set("Content-Length", strconv.FormatInt(object.Size, 10))

	_, err = io.Copy(c.Resp, r)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		log.Error("Failed to copy object file: %v", err)
		return
	}
	c.Status(http.StatusOK)
}

// PUT /{owner}/{repo}.git/info/lfs/object/basic/{oid}
func serveBasicUpload(c *context.Context, repo *db.Repository, oid lfsutil.OID) {
	err := db.LFS.CreateObject(repo.ID, oid, c.Req.Request.Body, lfsutil.StorageLocal)
	if err != nil {
		internalServerError(c.Resp)
		log.Error("Failed to create object [repo_id: %d, oid: %s]: %v", repo.ID, oid, err)
		return
	}
	c.Status(http.StatusOK)

	log.Trace("[LFS] Object created %q", oid)
}

// POST /{owner}/{repo}.git/info/lfs/object/basic/verify
func serveBasicVerify(c *context.Context, repo *db.Repository) {
	var request basicVerifyRequest
	defer c.Req.Request.Body.Close()
	err := json.NewDecoder(c.Req.Request.Body).Decode(&request)
	if err != nil {
		responseJSON(c.Resp, http.StatusBadRequest, responseError{
			Message: strutil.ToUpperFirst(err.Error()),
		})
		return
	}

	if !lfsutil.ValidOID(request.Oid) {
		responseJSON(c.Resp, http.StatusBadRequest, responseError{
			Message: "Invalid oid",
		})
		return
	}

	object, err := db.LFS.GetObjectByOID(repo.ID, lfsutil.OID(request.Oid))
	if err != nil {
		if db.IsErrLFSObjectNotExist(err) {
			responseJSON(c.Resp, http.StatusNotFound, responseError{
				Message: "Object does not exist",
			})
		} else {
			internalServerError(c.Resp)
			log.Error("Failed to get object [repo_id: %d, oid: %s]: %v", repo.ID, request.Oid, err)
		}
		return
	}

	if object.Size != request.Size {
		responseJSON(c.Resp, http.StatusNotFound, responseError{
			Message: "Object size mismatch",
		})
		return
	}
	c.Status(http.StatusOK)
}

type basicVerifyRequest struct {
	Oid  lfsutil.OID `json:"oid"`
	Size int64       `json:"size"`
}
