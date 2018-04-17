package web

import (
	"flag"
	"net/http"

	"git.workshop21.ch/ewa/common/go/abraxas/logging"

	"git.workshop21.ch/workshop21/ba/operator/configuration"
)

func Serve(config *configuration.Config) {
	port := config.WebPort
	directory := config.WebPath
	flag.Parse()

	http.Handle("/", http.FileServer(http.Dir(directory)))

	logging.WithID("BA-OPERATOR-FILESERV-001").Printf("Serving %s on HTTP port: %s\n", directory, port)
	logging.WithID("BA-OPERATOR-FILESERV-FATAL").Errorln(http.ListenAndServe(":"+port, nil))
}
