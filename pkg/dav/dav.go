package dav

import (
	"github.com/G-Node/gogs/pkg/context"
	"golang.org/x/net/webdav"
)

func Dav(c *context.Context, handler *webdav.Handler) {
	
	c.Write([]byte("Hallo"))
	return
}
