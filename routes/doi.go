package routes

import (
	"github.com/G-Node/gogs/pkg/context"
	"github.com/G-Node/gogs/pkg/setting"
	"fmt"
	"net/http"
)

func RequestDoi(c *context.Context) {
	if !c.Repo.IsAdmin(){
		c.Status(http.StatusUnauthorized)
		return
	}

	token := c.GetCookie(setting.SessionConfig.CookieName)
	c.Redirect(fmt.Sprintf("https://doi.gin.g-node.org/?repo=%s&user=%s&token=%s", c.Repo.Repository.FullName(),
		c.User.Name, token))
}
