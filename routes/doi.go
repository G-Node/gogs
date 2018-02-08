package routes

import (
	"github.com/G-Node/gogs/pkg/context"
	"github.com/G-Node/gogs/pkg/setting"
	"fmt"
	"net/http"
	"github.com/G-Node/gin-doi/src"
	log "gopkg.in/clog.v1"

)

func RequestDoi(c *context.Context) {
	if !c.Repo.IsAdmin() {
		c.Status(http.StatusUnauthorized)
		return
	}
	token := c.GetCookie(setting.SessionConfig.CookieName)
	token, err := ginDoi.Encrypt([]byte(setting.Doi.DoiKey), token)
	if err != nil {
		log.Error(0, "Could not encrypt Secret key:%s", err)
		c.Status(http.StatusInternalServerError)
		return
	}
	url := fmt.Sprintf("%s/?repo=%s&user=%s&token=%s", setting.Doi.DoiUrl, c.Repo.Repository.FullName(),
		c.User.Name, token)
	c.Redirect(url)
}
