package util

import (
	"strconv"

	queue "git.workshop21.ch/workshop21/ba/operator/metric-queue"
	"github.com/montanaflynn/stats"
)

func FloatToStr(fv float64) string {
	return strconv.FormatFloat(fv, 'f', 2, 64)
}

func MappingToArray(dataArray []queue.MetricTupel, number int) stats.Float64Data {
	data := make(stats.Float64Data, number)
	for i := 0; i < number; i++ {
		data[(number-1)-i] = dataArray[len(dataArray)-1-i].Value
	}
	return data
}
