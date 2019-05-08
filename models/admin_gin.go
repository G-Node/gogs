package models

import (
	gannex "github.com/G-Node/go-annex"
	log "gopkg.in/clog.v1"
)

func annexUninit(path string) {
	log.Trace("Uninit annex at '%s'", path)
	if msg, err := gannex.AUInit(path); err != nil {
		log.Error(1, "uninit failed: %v (%s)", err, msg)
		// TODO: Set mode 777 on all files to allow removal
	}
}
