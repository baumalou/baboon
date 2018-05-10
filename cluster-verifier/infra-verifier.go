package verifier

import (
	queue "git.workshop21.ch/workshop21/ba/operator/metric-queue"
	"git.workshop21.ch/workshop21/ba/operator/model"
	stats "git.workshop21.ch/workshop21/ba/operator/statistics"
	"git.workshop21.ch/workshop21/ba/operator/util"
)

func verifyCPUUsage(queue *queue.MetricQueue, length int) (model.StatValues, error) {

	usage := queue.GetNNewestTupel(length)
	if len(usage) == 0 {
		return util.GetStatValuesEmpty("cpu"), nil
	}
	result := stats.Mean(usage, length)
	//max := stats.Max(usage, length)
	//min := stats.Min(usage, length)
	deviation := stats.Deviation(usage, length)
	status := model.HEALTHY
	if result > 85 {
		status = model.ERROR
	} else if result >= 50 && result <= 85 {
		status = model.DEGRADED
	}
	perc90, err := stats.GetNPercentile(usage, 90)
	if err != nil {
		return util.GetStatValuesValue("cpu", result, status), nil
	}
	if perc90 > 85 {
		status = model.ERROR
	} else if perc90 >= 50 {
		status = model.DEGRADED
	}
	return util.GetStatValuesAll("cpu", result, status, deviation, status, perc90), nil
}

func verifyCPUCoresUsage(queue *queue.MetricQueue, length int) (model.StatValues, error) {

	usage := queue.GetNNewestTupel(length)
	if len(usage) == 0 {
		return util.GetStatValuesEmpty("cores"), nil
	}
	result := stats.Mean(usage, length)
	//max := stats.Max(usage, length)
	//min := stats.Min(usage, length)
	deviation := stats.Deviation(usage, length)
	status := model.HEALTHY
	if result > 50 {
		status = model.ERROR
	} else if result >= 30 && result <= 50 {
		status = model.DEGRADED
	}
	perc90, err := stats.GetNPercentile(usage, 90)
	if err != nil {
		return util.GetStatValuesValue("cpu", result, status), nil
	}
	if perc90 > 50 {
		status = model.ERROR
	} else if perc90 >= 30 {
		status = model.DEGRADED
	}
	return util.GetStatValuesAll("cores", result, status, deviation, status, perc90), nil
}

func verifyMemUsage(queue *queue.MetricQueue, length int) (model.StatValues, error) {

	usage := queue.GetNNewestTupel(length)
	if len(usage) == 0 {
		return util.GetStatValuesEmpty("memory"), nil
	}
	result := stats.Mean(usage, length)
	//max := stats.Max(usage, length)
	//min := stats.Min(usage, length)
	//deviation := stats.Deviation(usage, length)
	status := model.HEALTHY
	if result > 80 {
		status = model.ERROR
	} else if result >= 60 && result <= 80 {
		status = model.DEGRADED
	}
	return util.GetStatValuesValue("memory", result, status), nil
}

func verifyNetworkUsage(transmit *queue.MetricQueue, length int) (model.StatValues, error) {
	data := transmit.GetNNewestTupel(length)
	if len(data) == 0 {
		return util.GetStatValuesEmpty("network"), nil
	}
	result := stats.Mean(data, length)
	//max := stats.Max(data, length)
	//min := stats.Min(data, length)
	//deviation := stats.Deviation(data, length)

	status := model.HEALTHY
	if result > 80 {
		status = model.ERROR
	} else if result >= 50 && result <= 80 {
		status = model.DEGRADED
	}
	return util.GetStatValuesValue("network", result, status), nil
}
