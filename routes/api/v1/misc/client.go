package misc

import (
	"github.com/G-Node/gogs/pkg/context"
	"github.com/G-Node/gogs/pkg/setting"
	"encoding/json"
	"net/http"
	log "gopkg.in/clog.v1"
	"fmt"
)

type CliCongig struct {
	RsaHostKey string
}
type ApiCliConfig struct {
	RSAKet  string
	Weburl  string
	Webport string
	SSHUrl  string
	SSHUser string
	SSHPort int
}
func ClientC(c *context.APIContext) {
	cf := ApiCliConfig{RSAKet: setting.CliConfig.RsaHostKey,
		Weburl: fmt.Sprintf("%s://%s", setting.Protocol, setting.Domain),
		Webport: setting.HTTPPort, SSHUrl: setting.SSH.Domain, SSHPort: setting.SSH.Port,
		SSHUser: setting.RunUser}
	data, err := json.Marshal(cf)
	if err != nil {
		log.Info("Copuld not determine client congig: %+v", err)
		c.WriteHeader(http.StatusInternalServerError)
		return
	}
	c.WriteHeader(http.StatusOK)
	c.Write(data)
}
