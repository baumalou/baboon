package monitoring

import (
	"errors"
	"strconv"
	"sync"
	"time"

	"git.workshop21.ch/go/abraxas/logging"
	verifier "git.workshop21.ch/workshop21/ba/operator/cluster-verifier"
	"git.workshop21.ch/workshop21/ba/operator/configuration"
	queue "git.workshop21.ch/workshop21/ba/operator/metric-queue"
	"git.workshop21.ch/workshop21/ba/operator/statistics"
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
	go VerifyClusterStatusRoutine()
	for {
		wg.Add(len(datasets))
		for _, endpoint := range config.Endpoints {
			now := int(time.Now().Unix())
			go monitorRoutineSecs(datasets[endpoint.Name].Queue, config, endpoint.Path, now, config.SampleInterval)
		}
		wg.Wait()
		//verifier.VerifyClusterStatus(datasets)
		// why call this function when another already defined function exists in local package?
		//VerifyClusterStatus()
		time.Sleep(10 * time.Second)
	}
}

func VerifyClusterStatusRoutine() {
	for {
		VerifyClusterStatus()
		time.Sleep(1 * time.Second)
	}
}
func VerifyClusterStatus() bool {

	status, warning, _, err := verifier.VerifyClusterStatus(datasets)
	if err != nil {
		logging.WithError("BA-OPERATOR-MONITOR-003", err).Fatalln("not able to determine Cluster state", err)
	}
	switch warning {
	case verifier.DEGRADED:
		logging.WithID("BA-OPERATOR-MONITOR-WARNING-DEGRADED-005").Println("Cluster is nearly Degraded: ", warning)
	case verifier.ERROR:
		logging.WithID("BA-OPERATOR-MONITOR-WARNING-ERROR-005").Println("Cluster is nearly in Error State: ", warning)
	}
	switch status {
	case verifier.HEALTHY:
		logging.WithID("BA-OPERATOR-MONITOR-HEALTHY-004").Println("Cluster is Healthy: ", status)
		return true
	case verifier.DEGRADED:
		logging.WithID("BA-OPERATOR-MONITOR-DEGRADED-004").Println("Cluster is Degraded: ", status)
		return false
	case verifier.ERROR:
		logging.WithID("BA-OPERATOR-MONITOR-ERROR-004").Println("Cluster is in Error State!!! : ", status)
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
		go createEndpointDataset(datasets, config, endpoint)
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
