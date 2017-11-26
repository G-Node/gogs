package gindex

import (
	"net/http"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/gogits/go-gogs-client"

	"encoding/json"
	"bytes"
	"net/http/httptest"
	"strings"
)

// Handler for Index requests
func IndexH(w http.ResponseWriter, r *http.Request, els *ElServer, rpath *string) {
	rbd := IndexRequest{}
	err := getParsedBody(r, &rbd)
	log.Debugf("got a indexing request:%+v", rbd)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	err = IndexRepoWithPath(fmt.Sprintf("%s/%s", *rpath, strings.ToLower(rbd.RepoPath)+".git"),
		"master", els, rbd.RepoID, rbd.RepoPath)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	return
}

// Handler for SearchBlobs requests
func SearchH(w http.ResponseWriter, r *http.Request, els *ElServer, gins *GinServer) {
	rbd := SearchRequest{}
	err := getParsedBody(r, &rbd)
	log.Debugf("got a search request:%+v", rbd)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	// Get repo ids from the gin server to which the user has access
	// we need to limit results to those
	repos := []gogs.Repository{}
	err = getParsedHttpCall(http.MethodGet, fmt.Sprintf("%s/api/v1/user/repos", gins.URL),
		nil, rbd.Token, rbd.CsrfT, &repos)
	if err != nil {
		log.Errorf("could not querry repos: %+v", err)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	// Get repos ids for public repos
	prepos := struct{ Data []gogs.Repository }{}
	err = getParsedHttpCall(http.MethodGet, fmt.Sprintf("%s/api/v1/repos/search/?limit=10000", gins.URL),
		nil, rbd.Token, rbd.CsrfT, &prepos)
	if err != nil {
		log.Errorf("could not querry public repos: %+v", err)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	repos = append(repos, prepos.Data...)

	repids := make([]string, len(repos))
	for c, repo := range repos {
		repids[c] = fmt.Sprintf("%d", repo.ID)
	}
	log.Debugf("Repod to search in:%+v", repids)
	// Lets search now
	rBlobs := [] BlobSResult{}
	err = searchBlobs(rbd.Querry, repids, els, &rBlobs)
	if err != nil {
		log.Warnf("could not search blobs:%+v", err)
	}
	rCommits := [] CommitSResult{}
	err = searchCommits(rbd.Querry, repids, els, &rCommits)
	if err != nil {
		log.Warnf("could not search commits:%+v", err)
	}
	data, err := json.Marshal(SearchResults{Blobs: rBlobs, Commits: rCommits})
	if err != nil {
		log.Debugf("Could not Masrschal search results")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

// Handler for Index requests
func ReIndexRepo(w http.ResponseWriter, r *http.Request, els *ElServer, rpath *string) {
	rbd := IndexRequest{}
	err := getParsedBody(r, &rbd)
	log.Debugf("got a indexing request:%+v", rbd)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	err = ReIndexRepoWithPath(fmt.Sprintf("%s/%s", *rpath, strings.ToLower(rbd.RepoPath)+".git"),
		"master", els, rbd.RepoID, rbd.RepoPath)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	return
}
func ReindexH(w http.ResponseWriter, r *http.Request, els *ElServer, gins *GinServer, rpath *string) {
	rbd := ReIndexRequest{}
	getParsedBody(r, &rbd)
	log.Debugf("got a reindex request:%+v", rbd)
	repos, err := findRepos(*rpath, &rbd, gins)
	if err != nil {
		log.Debugf("failed listing repositories: %+v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	for _, repo := range repos {
		rec := httptest.NewRecorder()
		ireq := IndexRequest{rbd.UserID, repo.FullName,
			fmt.Sprintf("%d", repo.ID)}
		data, _ := json.Marshal(ireq)
		req, _ := http.NewRequest(http.MethodPost, "/index", bytes.NewReader(data))
		ReIndexRepo(rec, req, els, rpath)
		if rec.Code != http.StatusOK {
			log.Debugf("Could not index %s,%d", repo.FullName, rec.Code)
		}
	}
	w.WriteHeader(http.StatusOK)
}

func searchCommits(querry string, okRepids []string, els *ElServer,
	result interface{}) error {
	commS, err := els.SearchCommits(querry, okRepids)
	if err != nil {
		return err
	}
	err = parseElResult(commS, &result)
	commS.Body.Close()
	if err != nil {
		return err
	}
	return nil
}

func searchBlobs(querry string, okRepids []string, els *ElServer,
	result interface{}) error {
	blobS, err := els.SearchBlobs(querry, okRepids)
	if err != nil {
		return err
	}
	err = parseElResult(blobS, &result)
	blobS.Body.Close()
	if err != nil {
		return err
	}
	return nil
}

func parseElResult(comS *http.Response, pRes interface{}) error {
	var res interface{}
	err := getParsedResponse(comS, &res)
	if err != nil {
		return err
	}
	// extract the somewhat nested search rersult
	if x, ok := res.(map[string](interface{})); ok {
		if y, ok := x["hits"].(map[string](interface{})); ok {
			err = map2struct(y["hits"], &pRes)
			return err
		}
	}
	return fmt.Errorf("could not extract elastic result")
}
