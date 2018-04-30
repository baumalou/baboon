package web

import (
	"flag"
	"net/http"
	"strconv"
	"strings"
	"sync"

	"git.workshop21.ch/workshop21/ba/operator/fio-go"

	"git.workshop21.ch/go/abraxas/logging"
	"github.com/asaskevich/govalidator"
	"github.com/gorilla/mux"

	verifier "git.workshop21.ch/workshop21/ba/operator/cluster-verifier"
	"git.workshop21.ch/workshop21/ba/operator/configuration"
	"git.workshop21.ch/workshop21/ba/operator/kubeclient"
	"git.workshop21.ch/workshop21/ba/operator/monitoring"
)

var mutex *sync.Mutex
var running bool
var kc *kubeclient.KubeClient
var config *configuration.Config

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
	router.PathPrefix("/images/").Handler(http.StripPrefix("/images/", http.FileServer(http.Dir(directory))))
	router.HandleFunc("/run/{mode}/{size}", RunSmall).Methods("GET")
	router.HandleFunc("/getqueue/{endpoint}", PrintQueue).Methods("GET")
	router.HandleFunc("/getClusterState/{time}", GetClusterState).Methods("GET")
	router.HandleFunc("/run/{mode}/{size}/{bsize}", RunSmall).Methods("GET")
	logging.WithID("BA-OPERATOR-FILESERV-001").Printf("Serving %s on HTTP port: %s\n", directory, port)
	logging.WithID("BA-OPERATOR-FILESERV-FATAL").Errorln(http.ListenAndServe(":"+port, router))
}

func RunSmall(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	blockSize := "notSet"
	if len(params["bsize"]) > 0 {
		blockSize = params["bsize"]
	}
	if running {
		w.Write([]byte("already process running"))
		return
	} else if monitoring.VerifyClusterStatus() && params["size"] == "small" {
		handleEndpoint(params["mode"], params["size"], blockSize, w)
	} else if monitoring.VerifyClusterStatus() && params["size"] == "medium" {
		handleEndpoint(params["mode"], params["size"], blockSize, w)
	} else if monitoring.VerifyClusterStatus() && params["size"] == "large" {
		handleEndpoint(params["mode"], params["size"], blockSize, w)
	} else if !monitoring.VerifyClusterStatus() {
		w.Write([]byte("cluster not ready to run fio"))
	} else {
		w.Write([]byte("wrong command to run fio!! command:" + params["size"] + " \rpossible commands: small, medium, large"))
	}

}

func runFio(size, mode, bsize string) {
	lockMutex()
	defer unlockMutex()
	fio.RunFioAndGenPlot(size, mode, bsize)
}

func PrintQueue(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	if len(params["endpoint"]) < 4 {
		w.Write([]byte("wrong endpoint.\nallowed endpoints: " + strings.Join(monitoring.GetEndpoints(), ",")))
		return
	}
	dataset, err := monitoring.GetDataset(params["endpoint"])
	if err != nil {
		w.Write([]byte("wrong endpoint.\nallowed endpoints: " + strings.Join(monitoring.GetEndpoints(), ",") + " " + err.Error()))
		return
	}
	w.Write([]byte(dataset.Queue.PrintQueue()))
}

func GetClusterState(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	if govalidator.IsInt(params["time"]) {
		seconds, err := strconv.Atoi(params["time"])
		if err != nil {
			w.Write([]byte("could not get int of time parameter"))
			return
		}
		state, err := verifier.GetClusterStatusLastNSeconds(seconds)
		if err != nil {
			w.Write([]byte("could not get status of cluster"))
			return
		}
		w.Write([]byte(state))
		return
	}
	w.Write([]byte("Param is not integer"))
	return

}

func handleEndpoint(mode, size, bsize string, w http.ResponseWriter) {
	var err error
	if mode == "seq" {
		w.Write([]byte(mode + " " + size + " started"))
		go runFio(size, mode, bsize)
	} else if mode == "rand" {
		w.Write([]byte(mode + " " + size + " started"))
		go runFio(size, mode, bsize)
	} else {
		w.Write([]byte("wrong mode.\nallowed modes: seq, rand"))
		return
	}
	kc, err = kubeclient.GetKubeClient(kc)
	if err != nil {
		w.Write([]byte("not able to get kubeclient " + err.Error()))
		return
	}
	config, err = configuration.ReadConfig(config)
	if err != nil {
		w.Write([]byte("not able to get configuration " + err.Error()))
		return
	}
	err = kc.KillOnePodOf(config.RookOSDSelector)
	if err != nil {
		w.Write([]byte("not able to kill a pod out of " + config.RookOSDSelector + err.Error()))

	}
	return
}
