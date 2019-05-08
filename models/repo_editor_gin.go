package models

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

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
