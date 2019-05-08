package models

import (
	"strings"

	gannex "github.com/G-Node/go-annex"
	log "gopkg.in/clog.v1"
)

func annexUninit(path string) {
	if msg, err := gannex.AUInit(path); err != nil {
		if strings.Contains(msg, "If you don't care about preserving the data") {
			log.Trace("Annex uninit Repo: %s", msg)
		} else {
			log.Error(1, "Could not annex uninit repo. Error: %s,%s", err, msg)
		}
	} else {
		log.Trace("Annex uninit Repo:%s", msg)
	}
}
