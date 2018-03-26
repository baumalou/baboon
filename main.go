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
	"github.com/bmizerany/perks/quantile"
)

func main() {
	logging.WithID("PERF-OP-000").Info("operator started")
	config, err := configuration.ReadConfig(nil)
	if err != nil {
		logging.WithID("PERF-OP-1").Fatal(err)
	}
	//testing purpose:
	var data = [2][650]int{}

	for {
		now := int(time.Now().Unix())
		data[0] = getMonitoringData(config, config.OSDS_UP_Endpoint, now, 1)
		data[1] = getMonitoringData(config, config.AVG_OSD_APPLY_LATENCY, now, 1)

		time.Sleep(60 * time.Minute)
		log.Println(data)
	}

}

func getMonitoringData(config *configuration.Config, endpoint string, timeStampTo, hoursInPast int) [650]int {

	result, err := getGrafanaResultset(config, endpoint, timeStampTo, hoursInPast)
	if err != nil {
		logging.WithError("PERF-OP-h9u349u43", err)
		log.Println(err)
		return [650]int{}
	}
	logging.WithID("PERF-OP-0h8943o483f4o8").Info(result.Status)

	// Compute the 50th, 90th, and 99th percentile.
	q := quantile.NewTargeted(0.50, 0.90, 0.99)
	res := [650]int{}
	for _, res := range result.Data.Result[0].Values {
		// tm := time.Unix(int64(res[0].(float64)), 0)
		// if err != nil {
		// 	panic(err)
		// }
		value, _ := strconv.ParseFloat(res[1].(string), 64)
		ts := int(res[0].(float64))
		res = append(res, ts)
		q.Insert(value)

	}

	fmt.Println("perc50:", q.Query(0.50))
	fmt.Println("perc90:", q.Query(0.90))
	fmt.Println("perc99:", q.Query(0.99))
	fmt.Println("count:", q.Count())

	return res

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
