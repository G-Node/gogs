package main

import (
	"github.com/docopt/docopt-go"
	"os"
	"github.com/G-Node/gin-dex/gindex"
	"net/http"
	log  "github.com/Sirupsen/logrus"
)

func main() {
	usage := `gin-dex.
Usage:
  gin-dex [--eladress=<eladress> --eluser=<eluser> --elpw=<elpw> --rpath=<rpath> --gin=<gin> --port=<port> --debug ]

Options:
  --eladress=<eladress>           Adress of the elastic server [default: http://localhost:9200]
  --eluser=<eluser>               Elastic user [default: elastic]
  --elpw=<elpw>                   Elastic password [default: changeme]
  --port=<port>                   Server port [default: 8099]
  --gin=<gin>                     Gin Server Adress [default: https://gin.g-node.org]
  --rpath=<rpath>                 Path to the repositories [default: /repos]
  --debug                         Whether debug messages shall be printed
 `
	args, err := docopt.Parse(usage, nil, true, "gin-dex0.1a", false)
	if err != nil {
		log.Printf("Error while parsing command line: %+v", err)
		os.Exit(-1)
	}
	uname := args["--eluser"].(string)
	pw := args["--elpw"].(string)
	els := gindex.NewElServer(args["--eladress"].(string), &uname, &pw)
	gin := &gindex.GinServer{URL: args["--gin"].(string)}
	rpath := args["--rpath"].(string)

	http.HandleFunc("/index", func(w http.ResponseWriter, r *http.Request) {
		gindex.IndexH(w, r, els, &rpath)
	})

	http.HandleFunc("/search", func(w http.ResponseWriter, r *http.Request) {
		gindex.SearchH(w, r, els, gin)
	})

	http.HandleFunc("/reindex", func(w http.ResponseWriter, r *http.Request) {
		gindex.ReindexH(w, r, els, gin, &rpath)
	})


	if args["--debug"].(bool) {
		log.SetLevel(log.DebugLevel)
		log.SetFormatter(&log.TextFormatter{ForceColors: true})
	}
	log.Fatal(http.ListenAndServe(":"+args["--port"].(string), nil))
}