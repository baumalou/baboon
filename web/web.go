package web

import (
	"flag"
	"net/http"
	"sync"

	"git.workshop21.ch/workshop21/ba/operator/fio-go"

	"git.workshop21.ch/ewa/common/go/abraxas/logging"
	"github.com/gorilla/mux"

	"git.workshop21.ch/workshop21/ba/operator/configuration"
	"git.workshop21.ch/workshop21/ba/operator/monitoring"
)

var mutex *sync.Mutex
var running bool

func getMutex() *sync.Mutex {
	if mutex == nil {
		mutex = &sync.Mutex{}
	}
	return mutex
}

func lockMutex() {
	running = true
	logging.WithID("BA-OPERATOR-MUTEX-001").Debug("lock mutex")
	getMutex().Lock()
}
func unlockMutex() {
	getMutex().Unlock()
	running = false
	logging.WithID("BA-OPERATOR-MUTEX-003").Debug("mutex unlocked")
}

func Serve(config *configuration.Config) {
	port := config.WebPort
	directory := config.WebPath
	flag.Parse()
	router := mux.NewRouter()
	router.Handle("/{rest}", http.StripPrefix("/", http.FileServer(http.Dir(directory))))
	router.HandleFunc("/run/{size}", RunSmall).Methods("GET")
	logging.WithID("BA-OPERATOR-FILESERV-001").Printf("Serving %s on HTTP port: %s\n", directory, port)
	logging.WithID("BA-OPERATOR-FILESERV-FATAL").Errorln(http.ListenAndServe(":"+port, router))
}

func RunSmall(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	if running {
		w.Write([]byte("already process running"))
		return
	}
	if monitoring.VerifyClusterStatus() && params["size"] == "small" {
		w.Write([]byte("small started"))
		go runSmallFio()
		//fio.FioGenPlot()
	} else {
		w.Write([]byte("cluster not ready to run small"))
	}

}

func runSmallFio() {
	lockMutex()
	defer unlockMutex()
	fio.RunSmall()
}
