package models

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	gannex "github.com/G-Node/go-annex"
	"github.com/G-Node/gogs/pkg/setting"
	log "gopkg.in/clog.v1"
)

// StartIndexing sends an indexing request to the configured indexing service
// for a repository.
func StartIndexing(user, owner *User, repo *Repository) {
	var ireq struct{ RepoID, RepoPath string }
	ireq.RepoID = fmt.Sprintf("%d", repo.ID)
	ireq.RepoPath = repo.FullName()
	data, err := json.Marshal(ireq)
	if err != nil {
		log.Trace("could not marshal index request :%+v", err)
		return
	}
	req, _ := http.NewRequest(http.MethodPost, setting.Search.IndexUrl, bytes.NewReader(data))
	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil || resp.StatusCode != http.StatusOK {
		log.Trace("Error doing index request:%+v", err)
		return
	}
}

func annexUninit(path string) {
	log.Trace("Uninit annex at '%s'", path)
	if msg, err := gannex.AUInit(path); err != nil {
		log.Error(1, "uninit failed: %v (%s)", err, msg)
		// TODO: Set mode 777 on all files to allow removal
	}
}

func annexAdd(path string) {
	log.Trace("Running annex add (with filesize filter) in '%s'", path)

	// Initialise annex in case it's a new repository
	if msg, err := gannex.AInit(path, "--version=7"); err != nil {
		log.Error(1, "Annex init failed: %v (%s)", err, msg)
		return
	}

	// Enable addunlocked for annex v7
	if msg, err := gannex.SetAddUnlocked(path); err != nil {
		log.Error(1, "Failed to set 'addunlocked' annex option: %v (%s)", err, msg)
	}

	// Set MD5 as default backend
	if msg, err := gannex.Md5(path); err != nil {
		log.Error(1, "Failed to set default backend to 'MD5': %v (%s)", err, msg)
	}

	sizefilterflag := fmt.Sprintf("--largerthan=%d", setting.Repository.Upload.AnexFileMinSize*gannex.MEGABYTE)
	if msg, err := gannex.Add(path, sizefilterflag); err != nil {
		log.Error(1, "Annex add failed with error: %v (%s)", err, msg)
	}
}
