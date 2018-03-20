package main

import (
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
	config, _ := configuration.ReadConfig(nil)
	for {
		now := int(time.Now().Unix())
		getMonitoringData(config, config.OSDS_UP_Endpoint, now, 24)
		getMonitoringData(config, config.AVG_OSD_APPLY_LATENCY, now, 24)
		log.Println("operator started")
		time.Sleep(100 * time.Minute)
	}

}

func getMonitoringData(config *configuration.Config, endpoint string, timeStampTo, hoursInPast int) {

	startTime := -time.Duration(hoursInPast) * time.Hour
	start := int(time.Now().Add(startTime).Unix())
	url := config.MonitoringHost + endpoint + "&start=" + strconv.Itoa(start) + "&end=" + strconv.Itoa(timeStampTo) + "&step=" + config.SamplingStepSize

	var bearer = "Bearer " + config.BearerToken
	req, err := http.NewRequest("GET", url, nil)
	req.Header.Add("authorization", bearer)
	client := http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)
	result := model.GrafanaResult{}
	err = json.Unmarshal(body, &result)
	if err != nil {
		logging.WithError("PERF-OP-h9u349u43", err)
	} else {
		logging.WithID("PERF-OP-0h8943o483f4o8").Info(result.Status)
	}
	// Compute the 50th, 90th, and 99th percentile.
	q := quantile.NewTargeted(0.50, 0.90, 0.99)
	for _, res := range result.Data.Result[0].Values {
		// tm := time.Unix(int64(res[0].(float64)), 0)
		// if err != nil {
		// 	panic(err)
		// }
		value, _ := strconv.ParseFloat(res[1].(string), 64)
		q.Insert(value)
	}

	fmt.Println("perc50:", q.Query(0.50))
	fmt.Println("perc90:", q.Query(0.90))
	fmt.Println("perc99:", q.Query(0.99))
	fmt.Println("count:", q.Count())

}
