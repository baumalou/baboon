package statistics

import (
	queue "git.workshop21.ch/workshop21/ba/operator/metric-queue"
	"git.workshop21.ch/workshop21/ba/operator/util"
	stats "github.com/montanaflynn/stats"
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
