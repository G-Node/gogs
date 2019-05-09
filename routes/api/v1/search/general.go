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
	if !c.IsLogged {
		c.Status(http.StatusUnauthorized)
		return
	}
	if !setting.Search.Do {
		c.Status(http.StatusNotImplemented)
		return
	}
	ireq := libgin.SearchRequest{Token: c.GetCookie(setting.SessionConfig.CookieName), UserID: c.User.ID,
		Query: c.Params("query"), CsrfT: c.GetCookie(setting.CSRFCookieName)}
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
