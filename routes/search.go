package routes

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/G-Node/gogs/pkg/context"
	"github.com/G-Node/gogs/pkg/setting"
	"github.com/G-Node/libgin/libgin"
	log "gopkg.in/clog.v1"
)

const (
	EXPLORE_DATA    = "explore/data"
	EXPLORE_COMMITS = "explore/commits"
)

func Search(c *context.Context, keywords string, sType int64) ([]byte, error) {
	if !setting.Search.Do {
		c.Status(http.StatusNotImplemented)
		return nil, fmt.Errorf("Extended search not implemented")
	}

	ireq := libgin.SearchRequest{Token: c.GetCookie(setting.SessionConfig.CookieName),
		Query: keywords, CsrfT: c.GetCookie(setting.CSRFCookieName), SType: sType, UserID: -10}
	if c.IsLogged {
		ireq.UserID = c.User.ID
	}

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
	c.Data["Keywords"] = keywords
	sType, err := strconv.ParseInt(c.Query("stype"), 10, 0)
	if err != nil {
		log.Error(2, "Search type not understood: %s", err.Error())
		sType = 0
	}
	data, err := Search(c, keywords, sType)
	if err != nil {
		c.Handle(http.StatusInternalServerError, "Could not query", err)
		return
	}

	res := libgin.SearchResults{}
	err = json.Unmarshal(data, &res)
	if err != nil {
		c.Handle(http.StatusInternalServerError, "Could not display result", err)
		return
	}
	c.Data["Blobs"] = res.Blobs
	c.Data["opsel"] = sType
	c.HTML(200, EXPLORE_DATA)
}

func ExploreCommits(c *context.Context) {
	c.Data["Title"] = c.Tr("explore")
	c.Data["PageIsExplore"] = true
	c.Data["PageIsExploreCommits"] = true

	keywords := c.Query("q")
	sType, err := strconv.ParseInt(c.Query("stype"), 10, 0)
	if err != nil {
		log.Error(2, "Search type not understood: %s", err.Error())
		sType = 0
	}
	data, err := Search(c, keywords, sType)

	if err != nil {
		c.Handle(http.StatusInternalServerError, "Could not query", err)
		return
	}

	res := libgin.SearchResults{}
	err = json.Unmarshal(data, &res)
	if err != nil {
		c.Handle(http.StatusInternalServerError, "Could not display result", err)
		return
	}
	c.Data["Commits"] = res.Commits
	c.HTML(200, EXPLORE_COMMITS)
}

type SearchRequest struct {
	Token  string
	CsrfT  string
	UserID int64
	Query  string
	SType  int64
}
