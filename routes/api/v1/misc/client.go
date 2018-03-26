package misc

import (
	"github.com/G-Node/gogs/pkg/context"
	"github.com/G-Node/gogs/pkg/setting"
	"encoding/json"
	"net/http"
	log "gopkg.in/clog.v1"
)

type CliCongig struct {
	RsaHostKey string
}

func ClientC(c *context.APIContext) {
	data, err := json.Marshal(setting.CliConfig)
	if err != nil {
		log.Info("Copuld not determine client congig: %+v", err)
		c.WriteHeader(http.StatusInternalServerError)
		return
	}
	c.WriteHeader(http.StatusOK)
	c.Write(data)
}
