package verifier

import (
	"fmt"
	"strconv"

	"time"

	"git.workshop21.ch/go/abraxas/logging"
	"git.workshop21.ch/workshop21/ba/operator/configuration"
	queue "git.workshop21.ch/workshop21/ba/operator/metric-queue"
	stats "git.workshop21.ch/workshop21/ba/operator/statistics"
	"git.workshop21.ch/workshop21/ba/operator/util"
)

type StatValues struct {
	Name        string
	Value       float64
	ValueStatus int
	DevValue    float64
	DevStatus   int
}

func getStatValues(name string, value float64, valueStatus int, devValue float64, devStatus int) StatValues {
	data := StatValues{}
	data.Name = name
	data.Value = value
	data.ValueStatus = valueStatus
	data.DevValue = devValue
	data.DevStatus = devStatus
	return data
}

func StatValuesToString(struc StatValues) string {
	if struc.DevValue != 0.00 {
		return struc.Name + ": " + util.FloatToStr(struc.Value) + " " + statusToStr(struc.ValueStatus) + " " + util.FloatToStr(struc.DevValue) + " " + statusToStr(struc.DevStatus)
	} else {
		return struc.Name + ": " + util.FloatToStr(struc.Value) + " " + statusToStr(struc.ValueStatus)
	}
}
func StatValuesArrayToString(struc []StatValues) string {
	ret := ""
	for i := 0; i < len(struc); i++ {
		ret += StatValuesToString(struc[i]) + "\n"
	}
	return ret
}

const (
	HEALTHY int = 1 + iota
	DEGRADED
	ERROR
)

var config *configuration.Config

func getConfig() (*configuration.Config, error) {
	var err error
	if config == nil {
		config, err = configuration.ReadConfig(config)
		return config, err
	}
	return config, nil
}

// VerifyClusterStatus func cluster
func VerifyClusterStatus(dataset map[string]queue.Dataset) (int, int, []StatValues, error) {
	config, err := getConfig()
	if err != nil {
		return 0, 0, nil, err
	}
	length := config.LenghtRecordsToVerify
	logging.WithID("BA-OPERATOR-VERIFIER-01").Info("verifier started")

	data := make([]StatValues, 12)
	iops, iopsStatus, iopsDev, iopsWarning, err := verifyIOPS(dataset["IOPS_write"].Queue, dataset["IOPS_read"].Queue, length)
	data[0] = getStatValues("iops", iops, iopsStatus, iopsDev, iopsWarning)
	logging.WithID("BA-OPERATOR-VERIFIER-08").Info("IOPS: " + util.FloatToStr(iops) + " " + statusToStr(iopsStatus) + " " + util.FloatToStr(iopsDev) + " " + statusToStr(iopsWarning))

	mon, monStatus, err := verifyMonitorCounts(dataset["Mon_quorum"].Queue, length)
	data[1] = getStatValues("mon", mon, monStatus, 0.00, 0)
	logging.WithID("BA-OPERATOR-VERIFIER-09").Info("MonCount: " + util.FloatToStr(mon) + " " + statusToStr(monStatus))

	commit, commitStatus, commitDev, commitWarning, err := verifyOSDCommitLatency(dataset["AvOSDcommlat"].Queue, length)
	data[2] = getStatValues("commit", commit, commitStatus, commitDev, commitWarning)
	logging.WithID("BA-OPERATOR-VERIFIER-10").Info("CommitLat: " + util.FloatToStr(commit) + " " + statusToStr(commitStatus) + " " + util.FloatToStr(commitDev) + " " + statusToStr(commitWarning))

	apply, applyStatus, applyDev, applyWarning, err := verifyOSDApplyLatency(dataset["AvOSDappllat"].Queue, length)
	data[3] = getStatValues("apply", apply, applyStatus, applyDev, applyWarning)
	logging.WithID("BA-OPERATOR-VERIFIER-12").Info("ApplyLat: " + util.FloatToStr(apply) + " " + statusToStr(applyStatus) + " " + util.FloatToStr(applyDev) + " " + statusToStr(applyWarning))

	health, healthStatus, err := verifyCephHealth(dataset["CEPH_health"].Queue, length)
	data[4] = getStatValues("health", health, healthStatus, 0.00, 0)
	logging.WithID("BA-OPERATOR-VERIFIER-12").Info("CEPHHealth: " + util.FloatToStr(health) + " " + statusToStr(healthStatus))

	orphan, orphanStatus, err := verifyOSDOrphan(dataset["OSDInQuorum"].Queue, dataset["OSD_UP"].Queue, length)
	data[5] = getStatValues("orphan", orphan, orphanStatus, 0.00, 0)
	logging.WithID("BA-OPERATOR-VERIFIER-13").Info("Orphan: " + util.FloatToStr(orphan) + " " + statusToStr(orphanStatus))

	down, downStatus, err := verifyOSDDown(dataset["OSD_UP"].Queue, dataset["OSDInQuorum"].Queue, length)
	data[6] = getStatValues("down", down, downStatus, 0.00, 0)
	logging.WithID("BA-OPERATOR-VERIFIER-17").Info("Down: " + util.FloatToStr(down) + " " + statusToStr(downStatus))

	// tpRead, tpReadStatus, tpReadDev, tpWarning, err := verifyTPRead(dataset["TPread"].Queue, length)
	// data[7] = getStatValues("throughput read", tpRead, tpReadStatus, tpReadDev, tpWarning)

	// tpWrite, tpWriteStatus, tpWriteDev, tpWarning, err := verifyTPWrite(dataset["TPread"].Queue, length)
	// data[10] = getStatValues("throughput write", tpWrite, tpWriteStatus, tpWriteDev, tpWarning)

	infraStatus, err := VerfiyInfrastructureStatus(&data, dataset, length)

	logging.WithID("BA-OPERATOR-VERIFIER-02").Info("verifier finished")

	status := HEALTHY
	if iopsStatus == ERROR || monStatus == ERROR || commitStatus == ERROR || applyStatus == ERROR || healthStatus == ERROR || orphanStatus == ERROR || downStatus == ERROR || infraStatus == ERROR {
		status = ERROR
	} else if iopsStatus == DEGRADED || monStatus == DEGRADED || commitStatus == DEGRADED || applyStatus == DEGRADED || healthStatus == DEGRADED || orphanStatus == DEGRADED || downStatus == DEGRADED || infraStatus == DEGRADED {
		status = DEGRADED
	}
	warning := HEALTHY
	if iopsWarning == ERROR || commitWarning == ERROR || applyWarning == ERROR {
		warning = ERROR
	} else if iopsWarning == DEGRADED || commitWarning == DEGRADED || applyWarning == DEGRADED {
		warning = DEGRADED
	}
	return status, warning, data, err
}

// VerfiyInfrastructureStatus func infrastructure
func VerfiyInfrastructureStatus(struc *[]StatValues, dataset map[string]queue.Dataset, length int) (int, error) {
	yellow := 0
	red := 0
	values := *struc

	cpu, cpuStatus, err := verifyCPUUsage(dataset["PercUsedCPU"].Queue, length)
	logging.WithID("BA-OPERATOR-VERIFIER-03").Info("CPUUsage: " + util.FloatToStr(cpu) + " " + statusToStr(cpuStatus))
	values[7] = getStatValues("cpu", cpu, cpuStatus, 0.00, 0)
	if cpuStatus == DEGRADED {
		yellow += 3
	} else if cpuStatus == ERROR {
		red += 3
	}

	cores, coresStatus, err := verifyCPUCoresUsage(dataset["CPUCoresUsed"].Queue, length)
	logging.WithID("BA-OPERATOR-VERIFIER-04").Info("CoresUsage: " + util.FloatToStr(cores) + " " + statusToStr(coresStatus))
	values[8] = getStatValues("cores", cores, coresStatus, 0.00, 0)
	if coresStatus == DEGRADED {
		yellow += 3
	} else if coresStatus == ERROR {
		red += 3
	}

	mem, memStatus, err := verifyMemUsage(dataset["UsePercOfMem"].Queue, length)
	logging.WithID("BA-OPERATOR-VERIFIER-05").Info("MemUsage: " + util.FloatToStr(mem) + " " + statusToStr(memStatus))
	values[9] = getStatValues("memory", mem, memStatus, 0.00, 0)
	if memStatus == DEGRADED {
		yellow += 3
	} else if memStatus == ERROR {
		red += 3
	}

	net, netStatus, err := verifyNetworkUsage(dataset["networkTransmit"].Queue, length)
	logging.WithID("BA-OPERATOR-VERIFIER-06").Info("NetUsage: " + util.FloatToStr(net) + " " + statusToStr(netStatus))
	values[10] = getStatValues("network", net, netStatus, 0.00, 0)
	if netStatus == DEGRADED {
		yellow += 2
	} else if netStatus == ERROR {
		red += 2
	}

	cap, capStatus, err := verifyCapUsage(dataset["Av_capacity"].Queue, length)
	logging.WithID("BA-OPERATOR-VERIFIER-07").Info("CapUsage: " + util.FloatToStr(cap) + " " + statusToStr(capStatus))
	values[11] = getStatValues("capacity", cap, capStatus, 0.00, 0)
	if capStatus == DEGRADED {
		yellow++
	} else if capStatus == ERROR {
		red++
	}

	daysRemainingCap := predictDaysToCapacitiyLimit(dataset["Av_capacity"].Queue, length)
	logging.WithID("BA-OPERATOR-VERIFIER-15").Info("Predicted Day until Memory: " + statusToStr(daysRemainingCap))

	if red >= 1 {
		return ERROR, err
	} else if yellow >= 2 {
		return DEGRADED, err
	} else {
		return HEALTHY, err
	}
}

func predictDaysToCapacitiyLimit(data *queue.MetricQueue, length int) int {
	cap := data.GetNNewestTupel(length)
	timestamp, err := stats.ForecastRegression(cap)
	if err != nil {
		logging.WithError("BA-OPERATOR-VERIFIER-16", err).Error(err)
		return 0
	}
	pred := time.Unix(int64(timestamp), 0)
	diff := time.Until(pred)
	return int(diff.Hours()/24) / 1000
}

func verifyIOPS(write *queue.MetricQueue, read *queue.MetricQueue, length int) (float64, int, float64, int, error) {
	writeDS := write.GetNNewestTupel(length)
	readDS := read.GetNNewestTupel(length)
	if len(writeDS) == 0 || len(readDS) == 0 {
		return 0, HEALTHY, 0, HEALTHY, nil
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
	status := HEALTHY
	devStatus := HEALTHY
	limitYellow := 6000.00
	limitRed := 14000.00

	if deviation > limitRed || perc75 > limitRed {
		status = ERROR
	} else if deviation >= limitYellow && deviation <= limitRed {
		status = DEGRADED
	}

	if result > limitRed || perc75 > limitRed {
		status = ERROR
	} else if (result >= limitYellow && result <= limitRed) || perc75 > limitRed {
		status = DEGRADED
	}
	return result, status, deviation, devStatus, nil
}

func verifyMonitorCounts(queue *queue.MetricQueue, length int) (float64, int, error) {
	data := queue.GetNNewestTupel(length)
	if len(data) == 0 {
		return 0, HEALTHY, nil
	}
	min := stats.Min(data, length)

	if min < 2 {
		return min, ERROR, nil

	} else if min < 3 {
		return min, DEGRADED, nil
	} else {
		return min, HEALTHY, nil
	}
}

func verifyTPRead(queue *queue.MetricQueue, length int) (float64, int, float64, int, error) {
	commit := queue.GetNNewestTupel(length)
	if len(commit) == 0 {
		return 0, HEALTHY, 0, HEALTHY, nil
	}
	mean := stats.Mean(commit, length)
	max := stats.Max(commit, length)
	_, perc := stats.GetQuantiles(commit, config)
	fmt.Println(commit, mean, max, perc)

	deviation := stats.Deviation(commit, length)
	deviation += mean

	status := HEALTHY
	devStatus := HEALTHY
	limitYellow := 10.00
	limitRed := 50.00

	if deviation > limitRed {
		devStatus = ERROR
	} else if deviation >= limitYellow && deviation <= limitRed {
		devStatus = DEGRADED
	}

	if perc[0.90] > limitRed {
		status = ERROR
	} else if perc[0.90] >= limitYellow && perc[0.90] <= limitRed {
		status = DEGRADED
	}
	return perc[0.90], status, deviation, devStatus, nil

}

func verifyTPWrite(queue *queue.MetricQueue, length int) (float64, int, float64, int, error) {
	commit := queue.GetNNewestTupel(length)
	if len(commit) == 0 {
		return 0, HEALTHY, 0, HEALTHY, nil
	}
	mean := stats.Mean(commit, length)
	max := stats.Max(commit, length)
	_, perc := stats.GetQuantiles(commit, config)
	fmt.Println(commit, mean, max, perc)

	deviation := stats.Deviation(commit, length)
	deviation += mean

	status := HEALTHY
	devStatus := HEALTHY
	limitYellow := 10.00
	limitRed := 50.00

	if deviation > limitRed {
		devStatus = ERROR
	} else if deviation >= limitYellow && deviation <= limitRed {
		devStatus = DEGRADED
	}

	if perc[0.90] > limitRed {
		status = ERROR
	} else if perc[0.90] >= limitYellow && perc[0.90] <= limitRed {
		status = DEGRADED
	}
	return perc[0.90], status, deviation, devStatus, nil

}

func verifyOSDCommitLatency(queue *queue.MetricQueue, length int) (float64, int, float64, int, error) {

	commit := queue.GetNNewestTupel(length)
	result := stats.Mean(commit, length)
	if len(commit) == 0 {
		return 0, HEALTHY, 0, HEALTHY, nil
	}
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
	if len(apply) == 0 {
		return 0, HEALTHY, 0, HEALTHY, nil
	}
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
	if len(health) == 0 {
		return 0, HEALTHY, nil
	}
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
	if len(usage) == 0 {
		return 0, HEALTHY, nil
	}
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
	if len(usage) == 0 {
		return 0, HEALTHY, nil
	}
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
	if len(usage) == 0 {
		return 0, HEALTHY, nil
	}
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
func verifyNetworkUsage(transmit *queue.MetricQueue, length int) (float64, int, error) {
	data := transmit.GetNNewestTupel(length)
	if len(data) == 0 {
		return 0, HEALTHY, nil
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
	if len(inDS) == 0 || len(upDS) == 0 {
		return 0, HEALTHY, nil
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

	if max > 1 {
		return max, ERROR, nil
	} else if max == 1 {
		return max, DEGRADED, nil
	} else {
		return max, HEALTHY, nil
	}
}
func verifyOSDDown(up *queue.MetricQueue, in *queue.MetricQueue, length int) (float64, int, error) {
	inDS := in.GetNNewestTupel(length)
	upDS := up.GetNNewestTupel(length)
	if len(inDS) == 0 || len(upDS) == 0 {
		return 0, HEALTHY, nil
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
	if len(usage) == 0 {
		return 0, HEALTHY, nil
	}
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
