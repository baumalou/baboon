package web

import (
	"flag"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"git.workshop21.ch/workshop21/ba/operator/fio-go"
	queue "git.workshop21.ch/workshop21/ba/operator/metric-queue"

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
var wg sync.WaitGroup

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
	router.HandleFunc("/run/{mode}/{component}", RunSmall).Methods("GET")
	router.HandleFunc("/getqueue/{endpoint}", PrintQueue).Methods("GET")
	router.HandleFunc("/getClusterState/{time}", GetClusterState).Methods("GET")
	router.HandleFunc("/run/{mode}/{component}/{bsize}", RunSmall).Methods("GET")
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
	} else if monitoring.VerifyClusterStatus() && strings.Contains("all,none,osd,mon", params["component"]) {
		handleEndpoint(params["mode"], params["component"], blockSize, w)
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

		datasets := map[string]queue.Dataset{}

		wg.Add(len(datasets))
		go getDataForSecs(&datasets, seconds)
		wg.Wait()

		_, _, data, err := verifier.VerifyClusterStatus(datasets)
		state := verifier.StatValuesArrayToString(data)

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

func getDataForSecs(datasets *map[string]queue.Dataset, secs int) {
	defer wg.Done()
	for _, endpoint := range config.Endpoints {
		monQueue := queue.NewMetricQueue()
		(*datasets)[endpoint.Name] = queue.Dataset{Queue: monQueue, Name: endpoint.Name}
		now := int(time.Now().Unix())
		monitoring.MonitorRoutineSecs((*datasets)[endpoint.Name].Queue, config, endpoint.Path, now, secs)
	}
}

func handleEndpoint(mode, component, bsize string, w http.ResponseWriter) {
	var err error
	if mode == "seq" {
		w.Write([]byte(mode + " " + bsize + " started"))
		go runFio("small", mode, bsize)
	} else if mode == "rand" {
		w.Write([]byte(mode + " " + bsize + " started"))
		go runFio("small", mode, bsize)
	} else {
		w.Write([]byte("wrong mode.\nallowed modes: seq, rand"))
		return
	}
	config, err = configuration.ReadConfig(config)
	if err != nil {
		w.Write([]byte("\nnot able to get configuration " + err.Error()))
		return
	}
	if strings.Contains("mon,all", component) {
		killPod(config.RookMonSelector, w)
	}
	if strings.Contains("osd,all", component) {
		killPod(config.RookOSDSelector, w)
	}

	return
}

func killPod(component string, w http.ResponseWriter) {
	var err error
	kc, err = kubeclient.GetKubeClient(kc)
	if err != nil {
		w.Write([]byte("\nnot able to get kubeclient " + err.Error()))
		return
	}
	err = kc.KillOnePodOf(component)
	if err != nil {
		w.Write([]byte("\nnot able to kill a pod out of " + component + err.Error()))

	}
}
