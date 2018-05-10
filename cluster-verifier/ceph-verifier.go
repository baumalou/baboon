package verifier

import (
	queue "git.workshop21.ch/workshop21/ba/operator/metric-queue"
	"git.workshop21.ch/workshop21/ba/operator/model"
	stats "git.workshop21.ch/workshop21/ba/operator/statistics"
	"git.workshop21.ch/workshop21/ba/operator/util"
)

func verifyIOPS(write *queue.MetricQueue, read *queue.MetricQueue, length int) (model.StatValues, error) {
	writeDS := write.GetNNewestTupel(length)
	readDS := read.GetNNewestTupel(length)
	if len(writeDS) == 0 || len(readDS) == 0 {
		return util.GetStatValuesEmpty("iops"), nil
	}

	data := make([]queue.MetricTupel, length)
	for i := 0; i < length; i++ {
		data[i].Timestamp = writeDS[i].Timestamp
		data[i].Value = writeDS[i].Value + readDS[i].Value
	}
	result := stats.Mean(data, length)
	deviation := stats.Deviation(data, length)
	deviation += result
	perc75, _ := stats.GetNPercentQuantile(data, 0.75) //Example how to use percentile
	status := model.HEALTHY
	devStatus := model.HEALTHY
	limitYellow := 6000.00
	limitRed := 14000.00

	if deviation > limitRed || perc75 > limitRed {
		status = model.ERROR
	} else if deviation >= limitYellow && deviation <= limitRed {
		status = model.DEGRADED
	}

	if result > limitRed || perc75 > limitRed {
		status = model.ERROR
	} else if (result >= limitYellow && result <= limitRed) || perc75 > limitRed {
		status = model.DEGRADED
	}

	return util.GetStatValuesDev("iops", result, status, deviation, devStatus), nil
}

func verifyMonitorCounts(queue *queue.MetricQueue, length int) (model.StatValues, error) {
	data := queue.GetNNewestTupel(length)
	if len(data) == 0 {
		return util.GetStatValuesEmpty("mon"), nil
	}
	min := stats.Min(data, length)

	mon := model.HEALTHY
	if min < 2 {
		mon = model.ERROR
	} else if min < 3 {
		mon = model.DEGRADED
	}

	return util.GetStatValuesValue("mon", min, mon), nil
}

func verifyOSDCommitLatency(queue *queue.MetricQueue, length int) (model.StatValues, error) {

	commit := queue.GetNNewestTupel(length)
	result := stats.Mean(commit, length)
	if len(commit) == 0 {
		return util.GetStatValuesEmpty("commit"), nil
	}
	deviation := stats.Deviation(commit, length)
	deviation += result

	status := model.HEALTHY
	devStatus := model.HEALTHY
	limitYellow := 10.00
	limitRed := 50.00

	if deviation > limitRed {
		status = model.ERROR
	} else if deviation >= limitYellow && deviation <= limitRed {
		status = model.DEGRADED
	}

	if result > limitRed {
		status = model.ERROR
	} else if result >= limitYellow && result <= limitRed {
		status = model.DEGRADED
	}

	return util.GetStatValuesDev("commit", result, status, deviation, devStatus), nil
}

func verifyOSDApplyLatency(queue *queue.MetricQueue, length int) (model.StatValues, error) {
	apply := queue.GetNNewestTupel(length)
	if len(apply) == 0 {
		return util.GetStatValuesEmpty("apply"), nil
	}
	result := stats.Mean(apply, length)
	deviation := stats.Deviation(apply, length)
	deviation += result

	status := model.HEALTHY
	devStatus := model.HEALTHY
	limitYellow := 10.00
	limitRed := 50.00

	if deviation > limitRed {
		status = model.ERROR
	} else if deviation >= limitYellow && deviation <= limitRed {
		status = model.DEGRADED
	}

	if result > limitRed {
		status = model.ERROR
	} else if result >= limitYellow && result <= limitRed {
		status = model.DEGRADED
	}

	return util.GetStatValuesDev("apply", result, status, deviation, devStatus), nil
}

func verifyCephHealth(queue *queue.MetricQueue, length int) (model.StatValues, error) {
	health := queue.GetNNewestTupel(length)
	if len(health) == 0 {
		return util.GetStatValuesEmpty("health"), nil
	}
	max := stats.Max(health, length)

	res := model.HEALTHY
	if max == 2 {
		res = model.ERROR
	} else if max == 1 {
		res = model.DEGRADED
	}

	return util.GetStatValuesValue("health", max, res), nil
}

func verifyOSDOrphan(in *queue.MetricQueue, up *queue.MetricQueue, length int) (model.StatValues, error) {
	inDS := in.GetNNewestTupel(length)
	upDS := up.GetNNewestTupel(length)
	if len(inDS) == 0 || len(upDS) == 0 {
		return util.GetStatValuesEmpty("orphan"), nil
	}
	data := make([]queue.MetricTupel, length)
	for i := 0; i < length; i++ {
		data[i].Timestamp = upDS[i].Timestamp
		data[i].Value = upDS[i].Value - inDS[i].Value
		if data[i].Value < 0 {
			data[i].Value = 0
		}
	}
	//result := stats.Mean(data, length)
	max := stats.Max(data, length)
	//min := stats.Min(data, length)
	//deviation := stats.Deviation(data, length)

	status := model.HEALTHY
	if max > 1 {
		status = model.ERROR
	} else if max == 1 {
		status = model.DEGRADED
	}

	return util.GetStatValuesValue("orphan", max, status), nil
}
func verifyOSDDown(up *queue.MetricQueue, in *queue.MetricQueue, length int) (model.StatValues, error) {
	inDS := in.GetNNewestTupel(length)
	upDS := up.GetNNewestTupel(length)
	if len(inDS) == 0 || len(upDS) == 0 {
		return util.GetStatValuesEmpty("down"), nil
	}
	data := make([]queue.MetricTupel, length)
	for i := 0; i < length; i++ {
		data[i].Timestamp = upDS[i].Timestamp
		data[i].Value = inDS[i].Value - upDS[i].Value
		if data[i].Value < 0 {
			data[i].Value = 0
		}
	}
	//result := stats.Mean(data, length)
	max := stats.Max(data, length)
	//min := stats.Min(data, length)
	//deviation := stats.Deviation(data, length)
	status := model.HEALTHY
	if max > 1 {
		status = model.ERROR
	} else if max == 1 {
		status = model.DEGRADED
	}

	return util.GetStatValuesValue("down", max, status), nil
}

func verifyCapUsage(queue *queue.MetricQueue, length int) (model.StatValues, error) {

	usage := queue.GetNNewestTupel(length)
	if len(usage) == 0 {
		return util.GetStatValuesEmpty("capacity"), nil
	}
	result := stats.Mean(usage, length)
	//max := stats.Max(usage, length)
	//min := stats.Min(usage, length)
	//deviation := stats.Deviation(usage, length)

	status := model.HEALTHY
	if result > 80 {
		status = model.ERROR
	} else if result >= 10 && result <= 80 {
		status = model.DEGRADED
	}
	return util.GetStatValuesValue("capacity", result, status), nil
}

func verifyPG(queue *queue.MetricQueue, length int, metric string) (model.StatValues, error) {
	usage := queue.GetNNewestTupel(length)
	if len(usage) == 0 {
		return util.GetStatValuesEmpty(metric), nil
	}
	result := stats.Max(usage, length)
	status := model.HEALTHY
	if result > 10 {
		status = model.ERROR
	} else if result >= 1 {
		status = model.DEGRADED
	}
	return util.GetStatValuesValue("capacity", result, status), nil
}
