package search

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/G-Node/gogs/pkg/context"
	"github.com/G-Node/gogs/pkg/setting"
	"github.com/G-Node/libgin/libgin"
)

func Search(c *context.APIContext) {
	// TODO: API calls shouldn't use cookie (see https://github.com/G-Node/gin-dex/issues/2)
	if !c.IsLogged {
		c.Status(http.StatusUnauthorized)
		return
	}
	if setting.Search.SearchURL == "" {
		c.Status(http.StatusNotImplemented)
		return
	}
	ireq := libgin.SearchRequest{}
	data, err := json.Marshal(ireq)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}
	req, _ := http.NewRequest("Post", setting.Search.SearchURL, bytes.NewReader(data))
	cl := http.Client{}
	resp, err := cl.Do(req)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}
	data, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}
	c.Write(data)
}
