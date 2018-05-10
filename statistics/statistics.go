package statistics

import (
	queue "git.workshop21.ch/workshop21/ba/operator/metric-queue"
	"git.workshop21.ch/workshop21/ba/operator/util"
	stats "github.com/montanaflynn/stats"
	"github.com/sajari/regression"
)

// Min function
func Min(dataArray []queue.MetricTupel, number int) float64 {
	data := util.MappingToArray(dataArray, number)
	min, _ := stats.Min(data)
	return min
}

//Max function
func Max(dataArray []queue.MetricTupel, number int) float64 {
	data := util.MappingToArray(dataArray, number)
	max, _ := stats.Max(data)
	return max
}

//Mean function
func Mean(dataArray []queue.MetricTupel, number int) float64 {
	data := util.MappingToArray(dataArray, number)
	mean, _ := stats.Mean(data)
	return mean
}

// deviation
func Deviation(dataArray []queue.MetricTupel, number int) float64 {
	data := util.MappingToArray(dataArray, number)
	dev, _ := stats.MedianAbsoluteDeviationPopulation(data)
	return dev
}

// Forecast
func ForecastRegression(dataArray []queue.MetricTupel) (int, error) {
	r := new(regression.Regression)
	r.SetObserved("Timestamp")
	r.SetVar(0, "Capacity")
	for i := 0; i < len(dataArray); i++ {
		r.Train(
			regression.DataPoint(float64(dataArray[i].Timestamp), []float64{dataArray[i].Value}),
		)
	}
	r.Run()
	prediction, err := r.Predict([]float64{float64(80)})
	return int(prediction), err
}

// 75% percentile
func Percentile(dataArray []queue.MetricTupel, number int) float64 {
	data := util.MappingToArray(dataArray, number)
	perc, _ := stats.Percentile(data, 75)
	return perc
}
