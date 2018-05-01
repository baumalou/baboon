package statistics

import (
	"fmt"

	"git.workshop21.ch/go/abraxas/logging"
	"git.workshop21.ch/workshop21/ba/operator/configuration"
	queue "git.workshop21.ch/workshop21/ba/operator/metric-queue"
	"github.com/bmizerany/perks/quantile"
)

// GetQuantiles: Public Call for getQuantiles using a dataset
func GetQuantiles(dataset []queue.MetricTupel, config *configuration.Config) (string, map[float64]float64) {
	return getQuantiles(dataset, config)
}
func getQuantiles(dataset []queue.MetricTupel, config *configuration.Config) (string, map[float64]float64) {
	var quantileString string
	quantileMap := map[float64]float64{}
	q := quantile.NewTargeted(0.01, 0.10, 0.25, 0.50, 0.75, 0.80, 0.90, 0.95, 0.99, 0.999)
	for _, tupel := range dataset {
		q.Insert(tupel.Value)
	}

	for _, percentile := range config.Percentiles {
		perc := q.Query(percentile)
		quantileString = quantileString + fmt.Sprint(percentile, perc, "\n")
		quantileMap[percentile] = perc
	}
	quantileString = quantileString + fmt.Sprint("count:", q.Count())
	logging.WithID("BA-OPERATOR-QUANTILE-001").Println(quantileString)
	return quantileString, quantileMap
}
