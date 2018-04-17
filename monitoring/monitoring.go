package monitoring

import (
	"strconv"
	"time"

	"git.workshop21.ch/ewa/common/go/abraxas/logging"
	"git.workshop21.ch/workshop21/ba/operator/configuration"
	"git.workshop21.ch/workshop21/ba/operator/model"
	"github.com/bmizerany/perks/quantile"
)

func MonitorCluster(config *configuration.Config) {
	datasets := map[string]model.Dataset{}
	fillDataset(&datasets, config)
	for _, endpoint := range config.Endpoints {
		getQuantiles(datasets[endpoint.Name].Set, config)
	}
}

func fillDataset(datasets *map[string]model.Dataset, config *configuration.Config) {
	now := int(time.Now().Unix())
	for _, endpoint := range config.Endpoints {

		data := getMonitoringData(config, endpoint.Path, now, 3600)

		//queue := lang.NewQueue()
		// for timestamp, val := range data {
		// 	queue.Push(model.MetricTupel{Timestamp: timestamp, Value: val})
		// }
		(*datasets)[endpoint.Name] = model.Dataset{Set: data, Name: endpoint.Name}

		time.Sleep(100 * time.Millisecond)
	}
}

func getQuantiles(dataset []model.MetricTupel, config *configuration.Config) {
	q := quantile.NewTargeted(0.01, 0.10, 0.25, 0.50, 0.75, 0.80, 0.90, 0.95, 0.99)
	for _, tupel := range dataset {
		q.Insert(tupel.Value)
	}
	for _, percentile := range config.Percentiles {
		logging.WithID("BA-OPERATOR-QUANTILE-001").Println(percentile, q.Query(percentile))
	}
	logging.WithID("BA-OPERATOR-QUANTILE-COUNT").Println("count:", q.Count())
}

func getMonitoringData(config *configuration.Config, endpoint string, timeStampTo, hoursInPast int) []model.MetricTupel {

	result, err := getGrafanaResultset(config, endpoint, timeStampTo, hoursInPast)
	if err != nil {
		logging.WithError("PERF-OP-h9u349u43", err).Println(err)
		return nil
	}
	logging.WithID("PERF-OP-0h8943o483f4o8").Info(result.Status)

	// Compute the 50th, 90th, and 99th percentile.

	data := make([]model.MetricTupel, len(result.Data.Result))
	for _, res := range result.Data.Result[0].Values {
		// tm := time.Unix(int64(res[0].(float64)), 0)
		// if err != nil {
		// 	panic(err)
		// }
		value, _ := strconv.ParseFloat(res[1].(string), 64)
		ts := int(res[0].(float64))
		//fmt.Println(ts, "    ", value)
		data = append(data, model.MetricTupel{Timestamp: ts, Value: value})
		// if value == 0 {
		// 	log.Println(value, endpoint)
		// 	return data
		// }

	}

	return data

}
