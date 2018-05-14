package statistics

// Permission is hereby granted, free of charge, to any person obtaining a copy of this software and associated documentation files (the "Software"), to deal in the Software without restriction, including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so, subject to the following conditions:
// The above copyright notice and this permission notice shall be included in all copies or substantial portions of the Software.
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

import (
	"errors"
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

// GetNPercentQuantile returns the given percentile (float64) from an array of MetricTupel
func GetNPercentile(dataset []queue.MetricTupel, percentile float64) (float64, error) {
	if percentile > 1 {
		percentile = percentile / 100
	}
	q := quantile.NewTargeted(percentile)
	if len(dataset) < 1 {
		return 0, errors.New("dataset contains no Metric Tupel!\r")
	}
	for _, tupel := range dataset {
		q.Insert(tupel.Value)
	}
	return q.Query(percentile), nil
}
