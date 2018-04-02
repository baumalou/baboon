package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"time"

	"encoding/json"

	"git.workshop21.ch/ewa/common/go/abraxas/logging"
	"git.workshop21.ch/workshop21/ba/operator/configuration"
	"git.workshop21.ch/workshop21/ba/operator/model"
	"git.workshop21.ch/workshop21/ba/operator/storage"
	"github.com/bmizerany/perks/quantile"
)

func main() {
	logging.WithID("PERF-OP-000").Info("operator started")
	config, err := configuration.ReadConfig(nil)
	if err != nil {
		logging.WithID("PERF-OP-1").Fatal(err)
	}
	// err = testData(config)
	// if err != nil {
	// 	log.Fatal("went wrong!")
	// }
	//testing purpose:
	asStorage, err := storage.CreateClient(config)
	if err != nil {
		log.Println("not able to create as CLient")
		return
	}

	for {
		now := int(time.Now().Unix())
		osdUP := getMonitoringData(config, config.OSDS_UP_Endpoint, now, 1)
		//applyLatency := getMonitoringData(config, config.AVG_OSD_APPLY_LATENCY, now, 1)
		// IOPS_read := getMonitoringData(config, config.IOPS_read, now, 1)
		// IOPS_write := getMonitoringData(config, config.IOPS_write, now, 1)
		// Monitors_quorum := getMonitoringData(config, config.Monitors_quorum, now, 1)
		// Available_capacity := getMonitoringData(config, config.Available_capacity, now, 1)
		// AverageMonitorLatency := getMonitoringData(config, config.AverageMonitorLatency, now, 1)
		// Average_OSD_apply_latency := getMonitoringData(config, config.Average_OSD_apply_latency, now, 1)
		// Average_OSD_commit_latency := getMonitoringData(config, config.Average_OSD_commit_latency, now, 1)
		// Throughput_write := getMonitoringData(config, config.Throughput_write, now, 1)
		// Throughput_read := getMonitoringData(config, config.Throughput_read, now, 1)
		// CEPH_health := getMonitoringData(config, config.CEPH_health, now, 1)
		// OSD_Orphans := getMonitoringData(config, config.OSD_Orphans, now, 1)
		// Used_percent_of_cores := getMonitoringData(config, config.Used_percent_of_cores, now, 1)
		// Used_percent_of_memory := getMonitoringData(config, config.Used_percent_of_memory, now, 1)
		// network_usage := getMonitoringData(config, config.network_usage, now, 1)
		keys := make([]int, 0, len(osdUP))
		for ts, value := range osdUP {
			err = asStorage.WriteBin(ts, float64(ts), "Timestamp")
			if err != nil {
				log.Println(err.Error())
				return
			}
			err = asStorage.WriteBin(ts, value, "osdUP")
			if err != nil {
				log.Println(err.Error())
				return
			}
			keys = append(keys, ts)
		}
		time.Sleep(60 * time.Minute)

	}

}

func storeDataset(dataSet map[int]float64, keys []int, binName string, asStorage *storage.ASStorage) error {
	index := 0
	for _, value := range dataSet {
		err := asStorage.WriteBin(keys[index], value, binName)
		if err != nil {
			log.Println(err.Error())
			return err
		}
	}
	return nil
}

func testData(config *configuration.Config) error {
	log.Println(config.AerospikeHost)
	asStorage, err := storage.CreateClient(config)
	if err != nil {
		log.Println("not able to create as Client")
		return err
	}
	err = asStorage.WriteBin(1, 1, "test")
	if err != nil {
		log.Println(err.Error())
		return err
	}
	err = asStorage.WriteBin(1, 11, "test2")
	if err != nil {
		log.Println(err.Error())
		return err
	}
	return err
}

func getMonitoringData(config *configuration.Config, endpoint string, timeStampTo, hoursInPast int) map[int]float64 {

	result, err := getGrafanaResultset(config, endpoint, timeStampTo, hoursInPast)
	if err != nil {
		logging.WithError("PERF-OP-h9u349u43", err)
		log.Println(err)
		return nil
	}
	logging.WithID("PERF-OP-0h8943o483f4o8").Info(result.Status)

	// Compute the 50th, 90th, and 99th percentile.
	q := quantile.NewTargeted(0.50, 0.90, 0.99)
	data := make(map[int]float64)
	for _, res := range result.Data.Result[0].Values {
		// tm := time.Unix(int64(res[0].(float64)), 0)
		// if err != nil {
		// 	panic(err)
		// }
		value, _ := strconv.ParseFloat(res[1].(string), 64)
		ts := int(res[0].(float64))
		//fmt.Println(ts, "    ", value)
		data[ts] = value
		log.Println(value)
		q.Insert(value)

	}

	fmt.Println("perc50:", q.Query(0.50))
	fmt.Println("perc90:", q.Query(0.90))
	fmt.Println("perc99:", q.Query(0.99))
	fmt.Println("count:", q.Count())

	return data

}

func getGrafanaResultset(config *configuration.Config, endpoint string, timeStampTo, hoursInPast int) (model.GrafanaResult, error) {
	result := model.GrafanaResult{}
	startTime := -time.Duration(hoursInPast) * time.Hour
	start := int(time.Now().Add(startTime).Unix())
	url := config.MonitoringHost + endpoint + "&start=" + strconv.Itoa(start) + "&end=" + strconv.Itoa(timeStampTo) + "&step=" + config.SamplingStepSize
	logging.WithIDFields("PERF-OP-2").Info(url)
	var bearer = "Bearer " + config.BearerToken
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		logging.WithID("PERF-OP-2.974").Error(err)
		return result, err
	}
	req.Header.Add("authorization", bearer)
	client := http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		logging.WithID("PERF-OP-3").Error(resp, err)
		return result, err
	}
	if resp.StatusCode != 200 {
		logging.WithID("PERF-OP-h97843f7").Error(resp.Status, err)
		return result, errors.New("request failed with error: " + resp.Status)
	}
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	err = json.Unmarshal(body, &result)
	return result, err
}
