package verifier

import (
	"strconv"

	"git.workshop21.ch/go/abraxas/logging"
	queue "git.workshop21.ch/workshop21/ba/operator/metric-queue"
	stats "git.workshop21.ch/workshop21/ba/operator/statistics"
	"git.workshop21.ch/workshop21/ba/operator/util"
)

const (
	HEALTHY int = 1 + iota
	DEGRADED
	ERROR
)

// VerifyClusterStatus func cluster
func VerifyClusterStatus(dataset map[string]queue.Dataset) (int, error) {
	length := 20
	logging.WithID("BA-OPERATOR-VERIFIER-01").Info("verifier started")

	iops, iopsStatus, iopsDev, iopsWarning, err := verifyIOPS(dataset["IOPS_write"].Queue, dataset["IOPS_read"].Queue, length)
	logging.WithID("BA-OPERATOR-VERIFIER-08").Info("IOPS: " + util.FloatToStr(iops) + " " + statusToStr(iopsStatus) + " " + util.FloatToStr(iopsDev) + " " + statusToStr(iopsWarning))

	mon, monStatus, err := verifyMonitorCounts(dataset["Mon_quorum"].Queue, length)
	logging.WithID("BA-OPERATOR-VERIFIER-09").Info("MonCount: " + util.FloatToStr(mon) + " " + statusToStr(monStatus))

	commit, commitStatus, commitDev, commitWarning, err := verifyOSDCommitLatency(dataset["AvOSDcommlat"].Queue, length)
	logging.WithID("BA-OPERATOR-VERIFIER-10").Info("CommitLat: " + util.FloatToStr(commit) + " " + statusToStr(commitStatus) + " " + util.FloatToStr(commitDev) + " " + statusToStr(commitWarning))

	apply, applyStatus, applyDev, applyWarning, err := verifyOSDApplyLatency(dataset["AvOSDappllat"].Queue, length)
	logging.WithID("BA-OPERATOR-VERIFIER-12").Info("ApplyLat: " + util.FloatToStr(apply) + " " + statusToStr(applyStatus) + " " + util.FloatToStr(applyDev) + " " + statusToStr(applyWarning))

	health, healthStatus, err := verifyCephHealth(dataset["CEPH_health"].Queue, length)
	logging.WithID("BA-OPERATOR-VERIFIER-12").Info("CEPHHealth: " + util.FloatToStr(health) + " " + statusToStr(healthStatus))

	orphan, orphanStatus, err := verifyOSDOrphan(dataset["OSDInQuorum"].Queue, dataset["OSD_UP"].Queue, length)
	logging.WithID("BA-OPERATOR-VERIFIER-13").Info("Orphan: " + util.FloatToStr(orphan) + " " + statusToStr(orphanStatus))

	infraStatus, err := VerfiyInfrastructureStatus(dataset, length)

	logging.WithID("BA-OPERATOR-VERIFIER-02").Info("verifier finished")

	status := HEALTHY
	if iopsStatus == ERROR || monStatus == ERROR || commitStatus == ERROR || applyStatus == ERROR || healthStatus == ERROR || orphanStatus == ERROR || infraStatus == ERROR {
		status = ERROR
	} else if iopsStatus == DEGRADED || monStatus == DEGRADED || commitStatus == DEGRADED || applyStatus == DEGRADED || healthStatus == DEGRADED || orphanStatus == DEGRADED || infraStatus == DEGRADED {
		status = DEGRADED
	}

	logging.WithID("BA-OPERATOR-VERIFIER-14").Info("Result: " + statusToStr(status))
	return status, err

}

// VerfiyInfrastructureStatus func infrastructure
func VerfiyInfrastructureStatus(dataset map[string]queue.Dataset, length int) (int, error) {
	yellow := 0
	red := 0

	cpu, cpuStatus, err := verifyCPUUsage(dataset["PercUsedCPU"].Queue, length)
	logging.WithID("BA-OPERATOR-VERIFIER-03").Info("CPUUsage: " + util.FloatToStr(cpu) + " " + statusToStr(cpuStatus))
	if cpuStatus == DEGRADED {
		yellow += 3
	} else if cpuStatus == ERROR {
		red += 3
	}

	cores, coresStatus, err := verifyCPUCoresUsage(dataset["CPUCoresUsed"].Queue, length)
	logging.WithID("BA-OPERATOR-VERIFIER-04").Info("CoresUsage: " + util.FloatToStr(cores) + " " + statusToStr(coresStatus))
	if coresStatus == DEGRADED {
		yellow += 3
	} else if coresStatus == ERROR {
		red += 3
	}

	mem, memStatus, err := verifyMemUsage(dataset["UsePercOfMem"].Queue, length)
	logging.WithID("BA-OPERATOR-VERIFIER-05").Info("MemUsage: " + util.FloatToStr(mem) + " " + statusToStr(memStatus))
	if memStatus == DEGRADED {
		yellow += 3
	} else if memStatus == ERROR {
		red += 3
	}

	net, netStatus, err := verifyNetworkUsage(dataset["networkReceive"].Queue, dataset["networkSend"].Queue, length)
	logging.WithID("BA-OPERATOR-VERIFIER-06").Info("NetUsage: " + util.FloatToStr(net) + " " + statusToStr(netStatus))
	if netStatus == DEGRADED {
		yellow += 2
	} else if netStatus == ERROR {
		red += 2
	}

	cap, capStatus, err := verifyCapUsage(dataset["Av_capacity"].Queue, length)
	logging.WithID("BA-OPERATOR-VERIFIER-07").Info("CapUsage: " + util.FloatToStr(cap) + " " + statusToStr(capStatus))
	if capStatus == DEGRADED {
		yellow++
	} else if capStatus == ERROR {
		red++
	}

	if red >= 1 {
		return ERROR, err
	} else if yellow >= 2 {
		return DEGRADED, err
	} else {
		return HEALTHY, err
	}
}

func verifyIOPS(write *queue.MetricQueue, read *queue.MetricQueue, length int) (float64, int, float64, int, error) {
	writeDS := write.GetNNewestTupel(length)
	readDS := read.GetNNewestTupel(length)
	data := make([]queue.MetricTupel, length)
	for i := 0; i < length; i++ {
		data[i].Timestamp = writeDS[i].Timestamp
		data[i].Value = writeDS[i].Value + readDS[i].Value
	}
	result := stats.Mean(data, length)
	deviation := stats.Deviation(data, length)
	deviation += result

	status := HEALTHY
	devStatus := HEALTHY
	limitYellow := 6000.00
	limitRed := 14000.00

	if deviation > limitRed {
		status = ERROR
	} else if deviation >= limitYellow && deviation <= limitRed {
		status = DEGRADED
	}

	if result > limitRed {
		status = ERROR
	} else if result >= limitYellow && result <= limitRed {
		status = DEGRADED
	}
	return result, status, deviation, devStatus, nil
}

func verifyMonitorCounts(queue *queue.MetricQueue, length int) (float64, int, error) {
	data := queue.GetNNewestTupel(length)
	min := stats.Min(data, length)

	if min < 2 {
		return min, ERROR, nil
	} else if min < 3 {
		return min, DEGRADED, nil
	} else {
		return min, HEALTHY, nil
	}
}

func verifyOSDCommitLatency(queue *queue.MetricQueue, length int) (float64, int, float64, int, error) {

	commit := queue.GetNNewestTupel(length)
	result := stats.Mean(commit, length)
	deviation := stats.Deviation(commit, length)
	deviation += result

	status := HEALTHY
	devStatus := HEALTHY
	limitYellow := 10.00
	limitRed := 50.00

	if deviation > limitRed {
		status = ERROR
	} else if deviation >= limitYellow && deviation <= limitRed {
		status = DEGRADED
	}

	if result > limitRed {
		status = ERROR
	} else if result >= limitYellow && result <= limitRed {
		status = DEGRADED
	}
	return result, status, deviation, devStatus, nil
}

func verifyOSDApplyLatency(queue *queue.MetricQueue, length int) (float64, int, float64, int, error) {
	apply := queue.GetNNewestTupel(length)
	result := stats.Mean(apply, length)
	deviation := stats.Deviation(apply, length)
	deviation += result

	status := HEALTHY
	devStatus := HEALTHY
	limitYellow := 10.00
	limitRed := 50.00

	if deviation > limitRed {
		status = ERROR
	} else if deviation >= limitYellow && deviation <= limitRed {
		status = DEGRADED
	}

	if result > limitRed {
		status = ERROR
	} else if result >= limitYellow && result <= limitRed {
		status = DEGRADED
	}
	return result, status, deviation, devStatus, nil
}

func verifyCephHealth(queue *queue.MetricQueue, length int) (float64, int, error) {
	health := queue.GetNNewestTupel(length)
	max := stats.Max(health, length)

	if max == 2 {
		return max, ERROR, nil
	} else if max == 1 {
		return max, DEGRADED, nil
	} else {
		return max, HEALTHY, nil
	}
}

func verifyCPUUsage(queue *queue.MetricQueue, length int) (float64, int, error) {

	usage := queue.GetNNewestTupel(length)
	result := stats.Mean(usage, length)
	//max := stats.Max(usage, length)
	//min := stats.Min(usage, length)
	//deviation := stats.Deviation(usage, length)

	if result > 85 {
		return result, ERROR, nil
	} else if result >= 50 && result <= 85 {
		return result, DEGRADED, nil
	} else {
		return result, HEALTHY, nil
	}
}

func verifyCPUCoresUsage(queue *queue.MetricQueue, length int) (float64, int, error) {

	usage := queue.GetNNewestTupel(length)
	result := stats.Mean(usage, length)
	//max := stats.Max(usage, length)
	//min := stats.Min(usage, length)
	//deviation := stats.Deviation(usage, length)

	if result > 50 {
		return result, ERROR, nil
	} else if result >= 30 && result <= 50 {
		return result, DEGRADED, nil
	} else {
		return result, HEALTHY, nil
	}
}

func verifyMemUsage(queue *queue.MetricQueue, length int) (float64, int, error) {

	usage := queue.GetNNewestTupel(length)
	result := stats.Mean(usage, length)
	//max := stats.Max(usage, length)
	//min := stats.Min(usage, length)
	//deviation := stats.Deviation(usage, length)

	if result > 80 {
		return result, ERROR, nil
	} else if result >= 50 && result <= 80 {
		return result, DEGRADED, nil
	} else {
		return result, HEALTHY, nil
	}
}
func verifyNetworkUsage(rec *queue.MetricQueue, send *queue.MetricQueue, length int) (float64, int, error) {
	recDS := rec.GetNNewestTupel(length)
	sendDS := send.GetNNewestTupel(length)
	data := make([]queue.MetricTupel, length)
	for i := 0; i < length; i++ {
		data[i].Timestamp = recDS[i].Timestamp
		data[i].Value = (sendDS[i].Value + recDS[i].Value) / 1250000000 * 100
	}
	result := stats.Mean(data, length)
	//max := stats.Max(data, length)
	//min := stats.Min(data, length)
	//deviation := stats.Deviation(data, length)

	if result > 80 {
		return result, ERROR, nil
	} else if result >= 50 && result <= 80 {
		return result, DEGRADED, nil
	} else {
		return result, HEALTHY, nil
	}
}

func verifyOSDOrphan(in *queue.MetricQueue, up *queue.MetricQueue, length int) (float64, int, error) {
	inDS := in.GetNNewestTupel(length)
	upDS := up.GetNNewestTupel(length)
	data := make([]queue.MetricTupel, length)
	for i := 0; i < length; i++ {
		data[i].Timestamp = upDS[i].Timestamp
		data[i].Value = upDS[i].Value - inDS[i].Value
	}
	//result := stats.Mean(data, length)
	max := stats.Max(data, length)
	//min := stats.Min(data, length)
	//deviation := stats.Deviation(data, length)

	if max > 1 {
		return max, ERROR, nil
	} else if max == 1 {
		return max, DEGRADED, nil
	} else {
		return max, HEALTHY, nil
	}
}
func verifyCapUsage(queue *queue.MetricQueue, length int) (float64, int, error) {

	usage := queue.GetNNewestTupel(length)
	result := stats.Mean(usage, length)
	//max := stats.Max(usage, length)
	//min := stats.Min(usage, length)
	//deviation := stats.Deviation(usage, length)

	if result > 80 {
		return result, ERROR, nil
	} else if result >= 10 && result <= 80 {
		return result, DEGRADED, nil
	} else {
		return result, HEALTHY, nil
	}
}

func statusToStr(stat int) string {
	return strconv.Itoa(stat)
}
