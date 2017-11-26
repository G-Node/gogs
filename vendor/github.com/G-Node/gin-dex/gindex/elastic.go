package gindex

import (
	"bytes"
	"fmt"
	"net/http"

	"encoding/json"

	"io/ioutil"

	"github.com/G-Node/gig"
	log "github.com/Sirupsen/logrus"
)

type ElServer struct {
	adress   string
	uname    *string
	password *string
}

const (
	BLOB_INDEX   = "blobs"
	COMMIT_INDEX = "commits"
)

func NewElServer(adress string, uname, password *string) *ElServer {
	return &ElServer{adress: adress, uname: uname, password: password}
}

func (el *ElServer) Index(index, doctype string, data []byte, id gig.SHA1) (*http.Response, error) {
	adrr := fmt.Sprintf("%s/%s/%s/%s", el.adress, index, doctype, id.String())
	req, err := http.NewRequest("POST", adrr, bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	return el.elasticRequest(req)
}

func (el *ElServer) elasticRequest(req *http.Request) (*http.Response, error) {
	if el.uname != nil {
		req.SetBasicAuth(*el.uname, *el.password)
	}
	req.Header.Set("Content-Type", "application/json")
	cl := http.Client{}
	return cl.Do(req)
}

func (el *ElServer) HasCommit(index string, commitId gig.SHA1) (bool, error) {
	adrr := fmt.Sprintf("%s/commits/commit/%s", el.adress, commitId)
	return el.Has(adrr)
}

func (el *ElServer) HasBlob(index string, blobId gig.SHA1) (bool, error) {
	adrr := fmt.Sprintf("%s/blobs/blob/%s", el.adress, blobId)
	return el.Has(adrr)
}

func (el *ElServer) Has(adr string) (bool, error) {
	req, err := http.NewRequest("GET", adr, nil)
	if err != nil {
		return false, err
	}
	resp, err := el.elasticRequest(req)
	if err != nil {
		return false, err
	}
	bdy, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	var res struct{ Found bool }
	err = json.Unmarshal(bdy, &res)
	if err != nil {
		log.WithError(err)
		return false, err
	}
	return res.Found, nil
}

func (el *ElServer) search(querry, adrr string) (*http.Response, error) {
	req, err := http.NewRequest("POST", adrr, bytes.NewReader([]byte(querry)))
	if err != nil {
		log.Errorf("Could not form search query:%+v", err)
		log.Errorf("Formatted query was:%s", querry)
		return nil, err
	}
	return el.elasticRequest(req)
}

func (el *ElServer) SearchBlobs(querry string, okRepos []string) (*http.Response, error) {
	//implement the passing of the repo ids
	repos, err := json.Marshal(okRepos)
	if err != nil {
		log.Errorf("Could not marshal okRepos: %+v", err)
		return nil, err
	}
	formatted_querry := fmt.Sprintf(BLOB_QUERRY, querry, string(repos))
	adrr := fmt.Sprintf("%s/%s/_search", el.adress, BLOB_INDEX)
	return el.search(formatted_querry, adrr)
}

func (el *ElServer) SearchCommits(querry string, okRepos []string) (*http.Response, error) {
	//implement the passing of the repo ids
	repos, err := json.Marshal(okRepos)
	if err != nil {
		log.Errorf("Could not marshal okRepos: %+v", err)
		return nil, err
	}
	formatted_querry := fmt.Sprintf(COMMIT_QUERRY, querry, string(repos))
	adrr := fmt.Sprintf("%s/%s/_search", el.adress, COMMIT_INDEX)
	return el.search(formatted_querry, adrr)
}

var BLOB_QUERRY = `{
	"from" : 0, "size" : 20,
	  "_source": ["Oid","GinRepoName","FirstCommit","Path"],
	  "query": {
		"bool": {
		  "must": {
			"match": {
			  "_all": "%s"
			}
		  },
		  "filter": {
			"terms": {
			  "GinRepoId" : %s
			}
		  }
		}
	},
	"highlight" : {
		"fields" : [
			{"Content" : {
				"fragment_size" : 100,
				"number_of_fragments" : 10,
				"fragmenter": "span",
				"require_field_match":false,
				"pre_tags" : ["<b>"],
				"post_tags" : ["</b>"]
				}
			}
		]
	}
}`

var COMMIT_QUERRY = `{
	"from" : 0, "size" : 20,
	  "_source": ["Oid","GinRepoName","FirstCommit","Path"],
	  "query": {
		"bool": {
		  "must": {
			"match": {
			  "_all": "%s"
			}
		  },
		  "filter": {
			"terms": {
			  "GinRepoId" : %s
			}
		  }
		}
	},
	"highlight" : {
		"fields" : [
			{"Message" : {
				"fragment_size" : 50,
				"number_of_fragments" : 3,
				"fragmenter": "span",
				"require_field_match":false,
				"pre_tags" : ["<b>"],
				"post_tags" : ["</b>"]
				}
			},
			{"GinRepoName" : {
				"fragment_size" : 50,
				"number_of_fragments" : 3,
				"fragmenter": "span",
				"require_field_match":false,
				"pre_tags" : ["<b>"],
				"post_tags" : ["</b>"]
				}
			}
		]
	}
}`