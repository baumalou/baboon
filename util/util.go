package util

import (
	"math"
	"strconv"

	queue "git.workshop21.ch/workshop21/ba/operator/metric-queue"
	"git.workshop21.ch/workshop21/ba/operator/model"
	"github.com/montanaflynn/stats"
)

const (
	HEALTHY int = 1 + iota
	DEGRADED
	ERROR
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

func GetStatValuesAll(name string, value float64, valueStatus int, devValue float64, devStatus int, perc float64) model.StatValues {
	data := model.StatValues{}
	data.Name = name
	data.Value = value
	data.ValueStatus = valueStatus
	data.DevValue = devValue
	data.DevStatus = devStatus
	data.PercentileVal = perc
	return data
}
func GetStatValuesDev(name string, value float64, valueStatus int, devValue float64, devStatus int) model.StatValues {
	return GetStatValuesAll(name, value, valueStatus, devValue, devStatus, math.NaN())
}
func GetStatValuesValue(name string, value float64, valueStatus int) model.StatValues {
	return GetStatValuesDev(name, value, valueStatus, math.NaN(), model.HEALTHY)
}
func GetStatValuesEmpty(name string) model.StatValues {
	return GetStatValuesValue(name, math.NaN(), model.HEALTHY)
}

func StatValuesToString(struc model.StatValues) string {
	ret := struc.Name + ": " + FloatToStr(struc.Value) + " " + StatusToStr(struc.ValueStatus)

	if struc.DevValue != math.NaN() {
		ret = ret + " " + FloatToStr(struc.DevValue)
		if struc.PercentileVal != math.NaN() {
			ret = ret + " " + FloatToStr(struc.PercentileVal)
		}
		ret = ret + " " + StatusToStr(struc.DevStatus)
	}
	return ret
}

func StatValuesArrayToString(struc []model.StatValues) string {
	ret := ""
	for i := 0; i < len(struc); i++ {
		ret += StatValuesToString(struc[i]) + "\n"
	}
	return ret
}

func StatusToStr(stat int) string {
	return strconv.Itoa(stat)
}
