package route

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/G-Node/gogs/internal/context"
	"github.com/G-Node/gogs/internal/db"
	"github.com/G-Node/gogs/internal/setting"
	"github.com/G-Node/libgin/libgin"
	log "gopkg.in/clog.v1"
)

const (
	EXPLORE_DATA    = "explore/data"
	EXPLORE_COMMITS = "explore/commits"
)

type set map[int64]interface{}

func newset() set {
	return make(map[int64]interface{}, 0)
}

func (s set) add(item int64) {
	s[item] = nil
}

func (s set) contains(item int64) bool {
	_, yes := s[item]
	return yes
}

func (s set) remove(item int64) {
	delete(s, item)
}

func (s set) asSlice() []int64 {
	slice := make([]int64, len(s))
	idx := 0
	for item := range s {
		slice[idx] = item
		idx++
	}
	return slice
}

func collectSearchableRepoIDs(c *context.Context) ([]int64, error) {
	repoIDSet := newset()

	updateSet := func(repos []*db.Repository) {
		for _, r := range repos {
			repoIDSet.add(r.ID)
		}
	}

	if c.User != nil {
		ownRepos := c.User.Repos // user's own repositories
		updateSet(ownRepos)

		accessibleRepos, _ := c.User.GetAccessibleRepositories(0) // shared and org repos
		updateSet(accessibleRepos)
	}

	// Run a full repository search (with no keywords) to get public
	// repositories and then filter out the unlisted ones.
	repos, _, err := db.SearchRepositoryByName(&db.SearchRepoOptions{
		Keyword:  "",
		UserID:   c.UserID(),
		OrderBy:  "updated_unix DESC",
		Page:     0,
		PageSize: 0,
	})
	if err != nil {
		c.ServerError("SearchRepositoryByName", err)
		return nil, err
	}

	// If it's not unlisted, add it to the set
	// This will add public (listed) repositories
	for _, r := range repos {
		if !r.Unlisted {
			repoIDSet.add(r.ID)
		}
	}

	return repoIDSet.asSlice(), nil
}

func search(c *context.Context, keywords string, sType int) ([]byte, error) {
	if setting.Search.SearchURL == "" {
		log.Error(2, "Unable to perform search: SearchURL not configured")
		return nil, fmt.Errorf("Extended search not implemented")
	}

	key := []byte(setting.Search.Key)

	repoids, err := collectSearchableRepoIDs(c)
	if err != nil {
		log.Error(2, "Failed to collect searchable repository IDs: %v", err)
		return nil, err
	}
	searchdata := libgin.SearchRequest{Keywords: keywords, SType: sType, RepoIDs: repoids}

	data, err := json.Marshal(searchdata)
	if err != nil {
		log.Error(2, "Failed to marshal search request for gin-dex: %v", err)
		return nil, err
	}

	// encrypt query
	encdata, err := libgin.EncryptString(key, string(data))
	if err != nil {
		log.Error(2, "Failed to encrypt search data for gin-dex: %v", err)
		return nil, err
	}

	// Send query to gin-dex
	req, err := http.NewRequest("POST", setting.Search.SearchURL, strings.NewReader(encdata))
	if err != nil {
		log.Error(2, "Failed to build request for gin-dex: %v", err)
	}
	cl := http.Client{}
	resp, err := cl.Do(req)
	if err != nil {
		log.Error(2, "Failed to send request to gin-dex: %v", err)
		return nil, err
	}

	encrespdata, err := ioutil.ReadAll(resp.Body) // response is encrypted
	if err != nil {
		log.Error(2, "Failed to read response body from gin-dex: %v", err)
		return nil, err
	}

	// decrypt response
	respdata, err := libgin.DecryptString(key, string(encrespdata))
	if err != nil {
		log.Error(2, "Failed to decrypt response body form gin-dex: %v", err)
		return nil, err
	}

	return []byte(respdata), nil
}

// ExploreData handles the search box served at /explore/data
func ExploreData(c *context.Context) {
	keywords := c.Query("q")
	sType := c.QueryInt("stype") // non integer stype will return 0

	c.Data["Title"] = c.Tr("explore")
	c.Data["PageIsExplore"] = true
	c.Data["PageIsExploreData"] = true

	// send query data back even if the search fails or is aborted to fill in
	// the form on refresh
	c.Data["Keywords"] = keywords
	c.Data["opsel"] = sType

	res := libgin.SearchResults{}
	if keywords == "" {
		// no keywords submitted: don't search
		log.Trace("Loading empty data search page")
		c.Data["Blobs"] = res.Blobs
		c.HTML(200, EXPLORE_DATA)
		return
	}

	log.Trace("Searching data/blobs")
	data, err := search(c, keywords, sType)
	if err != nil {
		log.Error(2, "Query returned error: %v", err)
		c.Data["Blobs"] = res.Blobs
		c.HTML(200, EXPLORE_DATA)
		return
	}

	err = json.Unmarshal(data, &res)
	if err != nil {
		log.Error(2, "Failed to unmarshal response: %v", err)
		c.Data["Blobs"] = res.Blobs
		c.HTML(200, EXPLORE_DATA)
		return
	}
	c.Data["Blobs"] = res.Blobs
	c.HTML(200, EXPLORE_DATA)
}

// ExploreCommits handles the search box served at /explore/commits
func ExploreCommits(c *context.Context) {
	keywords := c.Query("q")
	sType := c.QueryInt("stype") // non integer stype will return 0

	c.Data["Title"] = c.Tr("explore")
	c.Data["PageIsExplore"] = true
	c.Data["PageIsExploreCommits"] = true

	// send query data back even if the search fails or is aborted to fill in
	// the form on refresh
	c.Data["Keywords"] = keywords
	c.Data["opsel"] = sType

	res := libgin.SearchResults{}
	if keywords == "" {
		log.Trace("Loading empty commit search page")
		// no keywords submitted: don't search
		c.Data["Commits"] = res.Commits
		c.HTML(200, EXPLORE_COMMITS)
		return
	}

	log.Trace("Searching commits")
	data, err := search(c, keywords, sType)

	if err != nil {
		log.Error(2, "Query returned error: %v", err)
		c.Data["Commits"] = res.Commits
		c.HTML(200, EXPLORE_COMMITS)
	}

	err = json.Unmarshal(data, &res)
	if err != nil {
		log.Error(2, "Failed to unmarshal response: %v", err)
		c.Data["Commits"] = res.Commits
		c.HTML(200, EXPLORE_COMMITS)
		return
	}
	c.Data["Commits"] = res.Commits
	c.HTML(200, EXPLORE_COMMITS)
}
