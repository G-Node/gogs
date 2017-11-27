package routes

import (
	"github.com/G-Node/gogs/pkg/context"
	"github.com/G-Node/gogs/pkg/setting"
	"fmt"
)

func RequestDoi(c *context.Context) {
	token := c.GetCookie(setting.SessionConfig.CookieName)
	c.Redirect(fmt.Sprintf("https://doi.gin.g-node.org/?repo=%s&user=%s&token=%s", c.Repo.Repository.FullName(),
		c.User.Name, token))
}
