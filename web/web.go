package web

import (
	"flag"
	"net/http"

	"git.workshop21.ch/workshop21/ba/operator/fio-go"

	"git.workshop21.ch/ewa/common/go/abraxas/logging"
	"github.com/gorilla/mux"

	"git.workshop21.ch/workshop21/ba/operator/configuration"
	"git.workshop21.ch/workshop21/ba/operator/monitoring"
)

func Serve(config *configuration.Config) {
	port := config.WebPort
	directory := config.WebPath
	flag.Parse()
	router := mux.NewRouter()
	router.Handle("/", http.FileServer(http.Dir(directory)))
	router.HandleFunc("/run/{size}", RunSmall).Methods("GET")
	logging.WithID("BA-OPERATOR-FILESERV-001").Printf("Serving %s on HTTP port: %s\n", directory, port)
	logging.WithID("BA-OPERATOR-FILESERV-FATAL").Errorln(http.ListenAndServe(":"+port, router))
}

func RunSmall(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)

	if monitoring.VerifyClusterStatus() && params["size"] == "small" {
		w.Write([]byte("small started"))
		go fio.RunSmall()
		//fio.FioGenPlot()
	} else {
		w.Write([]byte("cluster not ready to run small"))
	}

}
