package routes

import (
	"net/http"
	"net/url"

	"github.com/G-Node/gogs/pkg/context"
	"github.com/G-Node/gogs/pkg/setting"
	"github.com/G-Node/libgin/libgin"
	log "gopkg.in/clog.v1"
)

// RequestDOI sends a registration request to the configured DOI service
func RequestDOI(c *context.Context) {
	if !c.Repo.IsAdmin() {
		c.Status(http.StatusUnauthorized)
		return
	}

	repo := c.Repo.Repository.FullName()
	username := c.User.Name
	verification, err := libgin.EncryptURLString([]byte(setting.DOI.Key), repo+username)
	if err != nil {
		log.Error(2, "Could not encrypt secret key: %s", err)
		c.Status(http.StatusInternalServerError)
		return
	}
	doiurl, err := url.Parse(setting.DOI.URL + "/register") // TODO: Handle error by notifying admin email
	if err != nil {
		log.Error(2, "Failed to parse DOI URL: %s", setting.DOI.URL)
	}

	params := url.Values{}
	params.Add("repo", repo)
	params.Add("user", username)
	params.Add("verification", verification)
	doiurl.RawQuery = params.Encode()
	target, _ := url.PathUnescape(doiurl.String())
	log.Trace(target)
	c.RawRedirect(target)
}
