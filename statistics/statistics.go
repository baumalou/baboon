package statistics

import (
	queue "git.workshop21.ch/workshop21/ba/operator/metric-queue"
	stats "github.com/montanaflynn/stats"
)

// Min function
func Min(dataArray []queue.MetricTupel, number int) float64 {
	data := mappingToArray(dataArray, number)
	min, _ := stats.Min(data)
	return min
}

//Max function
func Max(dataArray []queue.MetricTupel, number int) float64 {
	data := mappingToArray(dataArray, number)
	max, _ := stats.Min(data)
	return max
}

//Mean function
func Mean(dataArray []queue.MetricTupel, number int) float64 {
	data := mappingToArray(dataArray, number)
	mean, _ := stats.Mean(data)
	return mean
}

// deviation
func Deviation(dataArray []queue.MetricTupel, number int) float64 {
	data := mappingToArray(dataArray, number)
	dev, _ := stats.MedianAbsoluteDeviationPopulation(data)
	return dev
}

func mappingToArray(dataArray []queue.MetricTupel, number int) stats.Float64Data {
	data := make(stats.Float64Data, 0, number)
	for i := 0; i < number; i++ {
		data[(number-1)-i] = dataArray[len(dataArray)-i].Value
	}
	return data
}1
