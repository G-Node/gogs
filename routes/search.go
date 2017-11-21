package routes

import (
	"github.com/G-Node/gogs/pkg/context"
	"net/http"
	"github.com/G-Node/gin-dex/gindex"
	"encoding/json"
	"bytes"
	"io/ioutil"
	"github.com/G-Node/gogs/pkg/setting"
	"fmt"
)

const (
	EXPLORE_DATA    = "explore/data"
	EXPLORE_COMMITS = "explore/commits"
)

func Search(c *context.Context, keywords string) ([]byte, error) {
	if ! c.IsLogged {
		c.Status(http.StatusUnauthorized)
		return nil, fmt.Errorf("User nor logged in")
	}
	if !setting.Search.Do {
		c.Status(http.StatusNotImplemented)
		return nil, fmt.Errorf("Extended search not implemented")
	}
	ireq := gindex.SearchRequest{Token: c.GetCookie(setting.SessionConfig.CookieName), UserID: c.User.ID,
		Querry: keywords, CsrfT: c.GetCookie(setting.CSRFCookieName)}
	data, err := json.Marshal(ireq)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return nil, err
	}
	req, _ := http.NewRequest("Post", setting.Search.SearchUrl, bytes.NewReader(data))
	cl := http.Client{}
	resp, err := cl.Do(req)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return nil, err
	}
	data, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return nil, err
	}
	return data, nil
}

func ExploreData(c *context.Context) {
	c.Data["Title"] = c.Tr("explore")
	c.Data["PageIsExplore"] = true
	c.Data["PageIsExploreData"] = true

	keywords := c.Query("q")
	data, err := Search(c, keywords)

	if err != nil {
		c.Handle(http.StatusInternalServerError, "Could nor querry", err)
		return
	}

	res := gindex.SearchResults{}
	err = json.Unmarshal(data, &res)
	if err != nil {
		c.Handle(http.StatusInternalServerError, "Could not display result", err)
		return
	}
	c.Data["Blobs"] = res.Blobs
	c.HTML(200, EXPLORE_DATA)
}

func ExploreCommits(c *context.Context) {
	c.Data["Title"] = c.Tr("explore")
	c.Data["PageIsExplore"] = true
	c.Data["PageIsExploreCommits"] = true

	keywords := c.Query("q")
	data, err := Search(c, keywords)

	if err != nil {
		c.Handle(http.StatusInternalServerError, "Could nor querry", err)
		return
	}

	res := gindex.SearchResults{}
	err = json.Unmarshal(data, &res)
	if err != nil {
		c.Handle(http.StatusInternalServerError, "Could not display result", err)
		return
	}
	c.Data["Commits"] = res.Commits
	c.HTML(200, EXPLORE_COMMITS)
}
