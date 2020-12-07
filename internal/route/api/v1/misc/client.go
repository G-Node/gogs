package misc

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/G-Node/gogs/internal/conf"
	"github.com/G-Node/gogs/internal/context"
	log "gopkg.in/clog.v1"
)

type CLICongig struct {
	RSAHostKey string
}
type APICLIConfig struct {
	RSAKet  string
	Weburl  string
	Webport string
	SSHUrl  string
	SSHUser string
	SSHPort int
}

func ClientC(c *context.APIContext) {
	cf := APICLIConfig{RSAKet: conf.CLIConfig.RSAHostKey,
		Weburl:  fmt.Sprintf("%s://%s", conf.Server.Protocol, conf.Server.Domain),
		Webport: conf.Server.HTTPPort, SSHUrl: conf.SSH.Domain, SSHPort: conf.SSH.Port,
		SSHUser: conf.App.RunUser}
	data, err := json.Marshal(cf)
	if err != nil {
		log.Info("Could not determine client configuration: %+v", err)
		c.WriteHeader(http.StatusInternalServerError)
		return
	}
	c.WriteHeader(http.StatusOK)
	_, _ = c.Write(data)
}
