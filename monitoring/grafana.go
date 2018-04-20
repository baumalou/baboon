package monitoring

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	"git.workshop21.ch/go/abraxas/logging"
	"git.workshop21.ch/workshop21/ba/operator/configuration"
	"git.workshop21.ch/workshop21/ba/operator/model"
)

func getGrafanaResultset(config *configuration.Config, endpoint string, timeStampTo, hoursInPast int) (model.GrafanaResult, error) {
	result := model.GrafanaResult{}
	startTime := -time.Duration(hoursInPast) * time.Second
	start := int(time.Now().Add(startTime).Unix())
	url := config.MonitoringHost + endpoint + "&start=" + strconv.Itoa(start) + "&end=" + strconv.Itoa(timeStampTo) + "&step=" + config.SamplingStepSize
	logging.WithIDFields("PERF-OP-2").Debug(url)
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
		return result, errors.New("request failed with error: " + resp.Status + "    " + url)
	}
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	err = json.Unmarshal(body, &result)
	return result, err
}
