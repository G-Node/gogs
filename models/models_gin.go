package models

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	gannex "github.com/G-Node/go-annex"
	"github.com/G-Node/gogs/pkg/setting"
	log "gopkg.in/clog.v1"
)

// StartIndexing sends an indexing request to the configured indexing service
// for a repository.
func StartIndexing(user, owner *User, repo *Repository) {
	if setting.Search.IndexURL == "" {
		log.Trace("Indexing not enabled")
		return
	}
	var ireq struct{ RepoID, RepoPath string }
	ireq.RepoID = fmt.Sprintf("%d", repo.ID)
	ireq.RepoPath = repo.FullName()
	data, err := json.Marshal(ireq)
	if err != nil {
		log.Trace("could not marshal index request :%+v", err)
		return
	}
	req, _ := http.NewRequest(http.MethodPost, setting.Search.IndexURL, bytes.NewReader(data))
	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil || resp.StatusCode != http.StatusOK {
		log.Trace("Error doing index request: %+v", err)
		return
	}
}

func annexUninit(path string) {
	// walker sets the permission for any file found to 0600, to allow deletion
	var mode os.FileMode
	walker := func(path string, info os.FileInfo, err error) error {
		if info == nil {
			return nil
		}

		mode = 0660
		if info.IsDir() {
			mode = 0770
		}

		if err := os.Chmod(path, mode); err != nil {
			log.Error(3, "failed to change permissions on '%s': %v", path, err)
		}
		return nil
	}

	log.Trace("Uninit annex at '%s'", path)
	if msg, err := gannex.Uninit(path); err != nil {
		log.Error(3, "uninit failed: %v (%s)", err, msg)
		if werr := filepath.Walk(path, walker); werr != nil {
			log.Error(3, "file permission change failed: %v", werr)
		}
	}
}

func annexAdd(path string) {
	log.Trace("Running annex add (with filesize filter) in '%s'", path)

	// Initialise annex in case it's a new repository
	if msg, err := gannex.Init(path, "--version=7"); err != nil {
		log.Error(1, "Annex init failed: %v (%s)", err, msg)
		return
	}

	// Enable addunlocked for annex v7
	if msg, err := gannex.SetAddUnlocked(path); err != nil {
		log.Error(1, "Failed to set 'addunlocked' annex option: %v (%s)", err, msg)
	}

	// Set MD5 as default backend
	if msg, err := gannex.MD5(path); err != nil {
		log.Error(1, "Failed to set default backend to 'MD5': %v (%s)", err, msg)
	}

	sizefilterflag := fmt.Sprintf("--largerthan=%d", setting.Repository.Upload.AnnexFileMinSize*gannex.MEGABYTE)
	if msg, err := gannex.Add(path, sizefilterflag); err != nil {
		log.Error(1, "Annex add failed with error: %v (%s)", err, msg)
	}
}

func annexSync(path string) error {
	log.Trace("Synchronising annexed data")
	if msg, err := gannex.ASync(path, "--content"); err != nil {
		// TODO: This will also DOWNLOAD content, which is unnecessary for a simple upload
		// TODO: Use gin-cli upload function instead
		log.Error(1, "Annex sync failed: %v (%s)", err, msg)
		return fmt.Errorf("git annex sync --content [%s]", path)
	}
	return nil
}
