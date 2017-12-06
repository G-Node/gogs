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
	"strconv"
	"github.com/Sirupsen/logrus"
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

	ireq := gindex.SearchRequest{Token: c.GetCookie(setting.SessionConfig.CookieName),
		Querry: keywords, CsrfT: c.GetCookie(setting.CSRFCookieName), SType: sType, UserID: -10}
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
	sType, err := strconv.ParseInt(c.Query("stype"), 10, 0)
	if err != nil {
		logrus.Errorf("Serach type not understood:%+v", err)
		sType = 0
	}
	data, err := Search(c, keywords, sType)
	if err != nil {
		c.Handle(http.StatusInternalServerError, "Could not querry", err)
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
	sType, err := strconv.ParseInt(c.Query("stype"), 10, 0)
	if err != nil {
		logrus.Errorf("Serach type not understood:%+v", err)
		sType = 0
	}
	data, err := Search(c, keywords, sType)

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
