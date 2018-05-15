package monitoring

import (
	"errors"
	"fmt"
	"strconv"
	"sync"
	"time"

	"git.workshop21.ch/go/abraxas/logging"
	verifier "git.workshop21.ch/workshop21/ba/operator/cluster-verifier"
	"git.workshop21.ch/workshop21/ba/operator/configuration"
	queue "git.workshop21.ch/workshop21/ba/operator/metric-queue"
	"git.workshop21.ch/workshop21/ba/operator/model"
	"git.workshop21.ch/workshop21/ba/operator/notifier"
	"git.workshop21.ch/workshop21/ba/operator/statistics"
	"git.workshop21.ch/workshop21/ba/operator/util"
)

var datasets map[string]queue.Dataset

var wg sync.WaitGroup

// GetDataset returns a COPY of a dataset
func GetDataset(endpoint string) (queue.Dataset, error) {
	value, ok := datasets[endpoint]
	if ok {
		return value, nil
	}
	return queue.Dataset{Queue: nil}, errors.New("Key not existent!")

}

// GetEndpoints returns a slice of keys as string
func GetEndpoints() []string {
	keys := make([]string, 0, len(datasets))
	for k := range datasets {
		keys = append(keys, k)
	}
	return keys
}

func MonitorCluster(config *configuration.Config) {
	datasets = map[string]queue.Dataset{}
	FillDataset(&datasets, config)
	for _, endpoint := range config.Endpoints {
		logging.WithID("BA-OPERATOR-MONITOR-" + endpoint.Name).Println("generating quantiles")
		statistics.GetQuantiles(datasets[endpoint.Name].Queue.Dataset, config)
	}
	go RecoveryWatcher()
	//go VerifyClusterStatusRoutine()
	for {
		wg.Add(len(datasets))
		for _, endpoint := range config.Endpoints {
			now := int(time.Now().Unix())
			go monitorRoutineSecs(datasets[endpoint.Name].Queue, config, endpoint.Path, now, config.SampleInterval)
		}
		wg.Wait()
		VerifyClusterStatus()
		//verifier.VerifyClusterStatus(datasets)
		// why call this function when another already defined function exists in local package?
		//VerifyClusterStatus()
		time.Sleep(time.Duration(config.SampleInterval) * time.Second)
	}
}

// func VerifyClusterStatusRoutine() {
// 	for {

// 		time.Sleep(config.SampleInterval * time.Second)
// 	}
// }
func VerifyClusterStatus() bool {
	notifier := notifier.GetNotifier()
	status, warning, vals, err := verifier.VerifyClusterStatus(datasets)
	if err != nil {
		logging.WithError("BA-OPERATOR-MONITOR-003", err).Fatalln("not able to determine Cluster state", err)
	}
	switch warning {
	case model.DEGRADED:
		notification := fmt.Sprintln("Cluster is nearly Degraded: ", warning)
		notifier.SendStatusNotification(util.StatValuesArrayToString(vals), notification)
		logging.WithID("BA-OPERATOR-MONITOR-WARNING-DEGRADED-005").Debug(notification)
	case model.ERROR:
		notification := fmt.Sprintln("Cluster is nearly in Error State: ", warning)
		notifier.SendStatusNotification(util.StatValuesArrayToString(vals), notification)
		logging.WithID("BA-OPERATOR-MONITOR-WARNING-ERROR-005").Debug(notification)

	}
	switch status {
	case model.HEALTHY:
		logging.WithID("BA-OPERATOR-MONITOR-HEALTHY-004").Debug("Cluster is Healthy: ", status)
		return true
	case model.DEGRADED:
		notification := fmt.Sprintln("Cluster is Degraded: ", status)
		notifier.SendStatusNotification(util.StatValuesArrayToString(vals), notification)
		logging.WithID("BA-OPERATOR-MONITOR-DEGRADED-004").Debug(notification)
		return false
	case model.ERROR:
		notification := fmt.Sprintln("Cluster is in Error State!!! : ", status)
		notifier.SendStatusNotification(util.StatValuesArrayToString(vals), notification)
		logging.WithID("BA-OPERATOR-MONITOR-ERROR-004").Debug(notification)
		return false
	}
	return false

}

func monitorRoutineSecs(mq *queue.MetricQueue, config *configuration.Config, endpoint string, timeTo int, secs int) {
	defer wg.Done()
	data := getMonitoringData(config, endpoint, timeTo, secs)
	mq.AddMonitoringTupelSliceToDataset(data)
	mq.Sort()
}

// ATTENTION!!!
// MonitorRoutineSecs inserts data to de provided queue
func MonitorRoutineSecs(mq *queue.MetricQueue, config *configuration.Config, endpoint string, timeTo int, secs int) {
	data := getMonitoringData(config, endpoint, timeTo, secs)
	mq.InsertMonitoringTupelInQueue(data)
	mq.Sort()
}

func FillDataset(datasets *map[string]queue.Dataset, config *configuration.Config) {
	wg.Add(len(config.Endpoints))
	for _, endpoint := range config.Endpoints {

		// go createEndpointDataset(datasets, config, endpoint)
		createEndpointDataset(datasets, config, endpoint)
	}
	wg.Wait()
}

func createEndpointDataset(datasets *map[string]queue.Dataset, config *configuration.Config, endpoint configuration.Endpoint) {
	defer wg.Done()
	now := int(time.Now().Unix())
	data := getMonitoringData(config, endpoint.Path, now, 3600)

	monQueue := queue.NewMetricQueue()
	monQueue.InsertMonitoringTupelInQueue(data)
	(*datasets)[endpoint.Name] = queue.Dataset{Queue: monQueue, Name: endpoint.Name}
}

func getMonitoringData(config *configuration.Config, endpoint string, timeStampTo, hoursInPast int) []queue.MetricTupel {

	result, err := GetGrafanaResultset(config, endpoint, timeStampTo, hoursInPast)
	if err != nil {
		logging.WithError("PERF-OP-h9u349u43", err).Println(err)
		return nil
	}
	logging.WithID("PERF-OP-0h8943o483f4o8").Debug(result.Status)

	// Compute the 50th, 90th, and 99th percentile.

	data := make([]queue.MetricTupel, len(result.Data.Result))
	if len(result.Data.Result) < 1 {
		logging.WithID("BA-OPERATOR-GETMONDATA").Println("no data received for", endpoint, "result: ", result.Status)
		return data
	}
	for _, res := range result.Data.Result[0].Values {
		// tm := time.Unix(int64(res[0].(float64)), 0)
		// if err != nil {
		// 	panic(err)
		// }
		value, _ := strconv.ParseFloat(res[1].(string), 64)
		ts := int(res[0].(float64))
		//fmt.Println(ts, "    ", value)
		data = append(data, queue.MetricTupel{Timestamp: ts, Value: value})
		// if value == 0 {
		// 	log.Println(value, endpoint)
		// 	return data
		// }

	}

	return data

}

// RecoveryWatcher Waits for the cluster to fail and then to recover
func RecoveryWatcher() {
	var timeElapsed int64
	config, err := configuration.ReadConfig(nil)
	notifier := notifier.GetNotifier()
	if err != nil {
		logging.WithID("RECOVERY-002").Error(err, err.Error(), "Not able to read Configuration")
		notifier.SendNotification("Not able to read Configuration")
	}
	timeElapsed, err = watchTimeTillClusterRecover(config)
	if err != nil {
		logging.WithID("RECOVERY-003").Error(err, err.Error())
		notifier.SendNotification("Error waiting for the Cluster to recover - " + err.Error())
	} else {
		logging.WithID("RECOVERY-004").Info("Cluster recovered after ", timeElapsed, " seconds")
	}
}

// run in chan!
func watchTimeTillClusterRecover(config *configuration.Config) (int64, error) {
	for _, endpoint := range config.Endpoints {

		if len(datasets[endpoint.Name].Queue.Dataset) < config.LenghtRecordsToVerify {
			return 0, errors.New("Not enough monitoring records")
		}
	}
	if waitTilClusterFails() {
		start := time.Now().Unix()
		if waitTilClusterRecovers() {
			return time.Now().Unix() - start, nil
		}
		return time.Now().Unix() - start, errors.New("Cluster did not recover in the expected time")

	}
	return 0, errors.New("Cluster did not fail in the expected time")

}

func waitTilClusterFails() bool {
	timer := time.Now().Unix()
	for !VerifyClusterStatus() && (time.Now().Unix())-timer < 300 {
		time.Sleep(1 * time.Second)
	}
	return !VerifyClusterStatus()
}

func waitTilClusterRecovers() bool {
	timer := time.Now().Unix()
	for VerifyClusterStatus() && (time.Now().Unix()-timer < 300) {
		time.Sleep(1 * time.Second)
	}
	return VerifyClusterStatus()
}
