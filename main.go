package main

import (
	"git.workshop21.ch/go/abraxas/logging"
	"git.workshop21.ch/workshop21/ba/operator/configuration"
	"git.workshop21.ch/workshop21/ba/operator/monitoring"
	"git.workshop21.ch/workshop21/ba/operator/web"
	"github.com/sirupsen/logrus"
)

func init() {
	logrus.SetLevel(logrus.DebugLevel)
}

func main() {
	logging.WithID("PERF-OP-000").Info("operator started")
	config, err := configuration.ReadConfig(nil)
	if err != nil {
		logging.WithID("PERF-OP-1").Fatal(err)
	}
	go monitoring.MonitorCluster(config)

	//runClassifier()
	//time.Sleep(60 * time.Second)
	//var keys []int

	web.Serve(config)
	/*
		for _, v := range datasets {
			keys = make([]int, 0, len(v))
			for ts := range v {
				keys = append(keys, ts)
				asStorage.WriteBin(ts, float64(ts), "Timestamp", "set")
			}
			break
		}

		for k, v := range datasets {
			err = storeDataset(v, keys, k, asStorage, "set")
			if err != nil {
				log.Println("some shit happened!!!!", k)
			}
		}
	*/
	// err = testData(config)
	// if err != nil {
	// 	log.Fatal("went wrong!")
	// }
	//testing purpose:
	/*
		asStorage, err := storage.CreateClient(config)
		if err != nil {
			log.Println("not able to create as CLient")
			return
		}
	*/

}
