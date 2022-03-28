package route

import (
	"encoding/json"
	"net/http"
	"net/url"

	"github.com/G-Node/libgin/libgin"
	"github.com/NII-DG/gogs/internal/conf"
	"github.com/NII-DG/gogs/internal/context"
	log "gopkg.in/clog.v1"
)

// RequestDOI sends a registration request to the configured DOI service
func RequestDOI(c *context.Context) {
	if !c.Repo.IsAdmin() {
		c.Status(http.StatusUnauthorized)
		return
	}

	username := c.User.Name
	realname := c.User.FullName
	repo := c.Repo.Repository.FullName()
	email := c.User.Email

	data := libgin.DOIRequestData{
		Username:   username,
		Realname:   realname,
		Repository: repo,
		Email:      email,
	}

	log.Trace("Encrypting data for DOI: %+v", data)
	dataj, _ := json.Marshal(data)
	regrequest, err := libgin.EncryptURLString([]byte(conf.DOI.Key), string(dataj))
	if err != nil {
		log.Error(2, "Could not encrypt secret key: %s", err)
		c.Status(http.StatusInternalServerError)
		return
	}
	doiurl, err := url.Parse(conf.DOI.URL + "/register") // TODO: Handle error by notifying admin email
	if err != nil {
		log.Error(2, "Failed to parse DOI URL: %s", conf.DOI.URL)
	}

	params := url.Values{}
	params.Add("regrequest", regrequest)
	doiurl.RawQuery = params.Encode()
	target, _ := url.PathUnescape(doiurl.String())
	log.Trace(target)
	c.RawRedirect(target)
}
