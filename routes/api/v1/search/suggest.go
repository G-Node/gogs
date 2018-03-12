package search

import (
	"net/http"
	"github.com/G-Node/gin-dex/gindex"
	"encoding/json"
	"bytes"
	"io/ioutil"
	"github.com/G-Node/gogs/pkg/context"
	"github.com/G-Node/gogs/pkg/setting"
)

func Suggest(c *context.APIContext) {
	if ! c.IsLogged {
		c.Status(http.StatusUnauthorized)
		return
	}
	if !setting.Search.Do {
		c.Status(http.StatusNotImplemented)
		return
	}
	ireq := gindex.SearchRequest{Token: c.GetCookie(setting.SessionConfig.CookieName), UserID: c.User.ID,
		Querry: c.Params("querry"), CsrfT: c.GetCookie(setting.CSRFCookieName), SType:gindex.SEARCH_SUGGEST}
	data, err := json.Marshal(ireq)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}
	req, _ := http.NewRequest("Post", setting.Search.SearchUrl, bytes.NewReader(data))
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
