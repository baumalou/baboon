package monitoring

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"git.workshop21.ch/go/abraxas/logging"
	verifier "git.workshop21.ch/workshop21/ba/operator/cluster-verifier"
	"git.workshop21.ch/workshop21/ba/operator/configuration"
	queue "git.workshop21.ch/workshop21/ba/operator/metric-queue"
	"git.workshop21.ch/workshop21/ba/operator/model"
	"git.workshop21.ch/workshop21/ba/operator/statistics"
	"git.workshop21.ch/workshop21/ba/operator/util"
)

var datasets map[string]queue.Dataset

var wg sync.WaitGroup

var notificationTimer NotificationTimer

type NotificationTimer struct {
	ErrorTimer    int64
	DegradedTimer int64
}

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
	notificationTimer = NotificationTimer{ErrorTimer: 0, DegradedTimer: 0}
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

	status, warning, vals, err := verifier.VerifyClusterStatus(datasets)
	if err != nil {
		logging.WithError("BA-OPERATOR-MONITOR-003", err).Fatalln("not able to determine Cluster state", err)
	}
	switch warning {
	case model.DEGRADED:
		notification := fmt.Sprintln("Cluster is nearly Degraded: ", warning)
		SendNOtification(util.StatValuesArrayToString(vals), notification)
		logging.WithID("BA-OPERATOR-MONITOR-WARNING-DEGRADED-005").Println(notification)
	case model.ERROR:
		notification := fmt.Sprintln("Cluster is nearly in Error State: ", warning)
		SendNOtification(util.StatValuesArrayToString(vals), notification)
		logging.WithID("BA-OPERATOR-MONITOR-WARNING-ERROR-005").Println(notification)

	}
	switch status {
	case model.HEALTHY:
		logging.WithID("BA-OPERATOR-MONITOR-HEALTHY-004").Println("Cluster is Healthy: ", status)
		return true
	case model.DEGRADED:
		notification := fmt.Sprintln("Cluster is Degraded: ", status)
		SendNOtification(util.StatValuesArrayToString(vals), notification)
		logging.WithID("BA-OPERATOR-MONITOR-DEGRADED-004").Println(notification)
		return false
	case model.ERROR:
		notification := fmt.Sprintln("Cluster is in Error State!!! : ", status)
		SendNOtification(util.StatValuesArrayToString(vals), notification)
		logging.WithID("BA-OPERATOR-MONITOR-ERROR-004").Println(notification)
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
func notificationNeedsToBeSent(notification string) bool {
	if strings.Contains(notification, "Degraded") {
		if time.Now().Unix()-notificationTimer.DegradedTimer > int64((30 * time.Minute).Seconds()) {
			notificationTimer.DegradedTimer = time.Now().Unix()
			return true
		}
		return false
	} else if strings.Contains(notification, "Error") {
		if time.Now().Unix()-notificationTimer.ErrorTimer > int64((30 * time.Minute).Seconds()) {
			notificationTimer.ErrorTimer = time.Now().Unix()
			return true
		}
		return false
	}
	return true

}
func SendNOtification(vals, notification string) {
	if !notificationNeedsToBeSent(notification) {
		return
	}
	notification = vals + notification
	url := "https://chat.workshop21.ch/hooks/5zhbybp88jgwp88zanu9j4751w"
	fmt.Println("URL:>", url)

	var jsonStr = []byte(`
		{
			"text": "` + notification + `"
		}`)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonStr))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	fmt.Println("response Status:", resp.Status)
	fmt.Println("response Headers:", resp.Header)
	body, _ := ioutil.ReadAll(resp.Body)
	fmt.Println("response Body:", string(body))

}
