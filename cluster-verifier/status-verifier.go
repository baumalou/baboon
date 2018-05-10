package verifier

import (
	"fmt"
	"math"

	"time"

	"git.workshop21.ch/go/abraxas/logging"
	"git.workshop21.ch/workshop21/ba/operator/configuration"
	queue "git.workshop21.ch/workshop21/ba/operator/metric-queue"
	"git.workshop21.ch/workshop21/ba/operator/model"
	stats "git.workshop21.ch/workshop21/ba/operator/statistics"
	"git.workshop21.ch/workshop21/ba/operator/util"
)

var nan = math.NaN()

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
func VerifyClusterStatus(dataset map[string]queue.Dataset) (int, int, []model.StatValues, error) {
	config, err := getConfig()
	if err != nil {
		return 0, 0, nil, err
	}
	length := config.LenghtRecordsToVerify
	logging.WithID("BA-OPERATOR-VERIFIER-01").Info("verifier started")

	var cephdata []model.StatValues
	err = VerifyCephStatus(&cephdata, dataset, length)

	var infradata []model.StatValues
	infraStatus, err := VerfiyInfrastructureStatus(&infradata, dataset, length)

	logging.WithID("BA-OPERATOR-VERIFIER-02").Info("verifier finished")

	status := model.HEALTHY
	warning := model.HEALTHY
	for i := 0; i < len(cephdata); i++ {
		if cephdata[i].ValueStatus > status {
			status = cephdata[i].ValueStatus
		}
		// possible because DevStatus is HEALTHY if irrelevant
		if cephdata[i].DevStatus > warning {
			warning = cephdata[i].DevStatus
		}
	}
	if infraStatus == model.ERROR {
		status = model.ERROR
	} else if status == model.HEALTHY && infraStatus == model.DEGRADED {
		status = model.DEGRADED
	}

	return status, warning, append(cephdata, infradata...), err
}

// VerifyCephStatus analyse ceph status
func VerifyCephStatus(struc *[]model.StatValues, dataset map[string]queue.Dataset, length int) error {

	iops, err := verifyIOPS(dataset["IOPS_write"].Queue, dataset["IOPS_read"].Queue, length)
	*struc = append(*struc, iops)
	logging.WithID("BA-OPERATOR-VERIFIER-08").Info(util.StatValuesToString(iops))

	mon, err := verifyMonitorCounts(dataset["Mon_quorum"].Queue, length)
	*struc = append(*struc, mon)
	logging.WithID("BA-OPERATOR-VERIFIER-09").Info(util.StatValuesToString(mon))

	commit, err := verifyOSDCommitLatency(dataset["AvOSDcommlat"].Queue, length)
	*struc = append(*struc, commit)
	logging.WithID("BA-OPERATOR-VERIFIER-10").Info(util.StatValuesToString(commit))

	apply, err := verifyOSDApplyLatency(dataset["AvOSDappllat"].Queue, length)
	*struc = append(*struc, apply)
	logging.WithID("BA-OPERATOR-VERIFIER-12").Info(util.StatValuesToString(apply))

	health, err := verifyCephHealth(dataset["CEPH_health"].Queue, length)
	*struc = append(*struc, health)
	logging.WithID("BA-OPERATOR-VERIFIER-12").Info(util.StatValuesToString(health))

	orphan, err := verifyOSDOrphan(dataset["OSDInQuorum"].Queue, dataset["OSD_UP"].Queue, length)
	*struc = append(*struc, orphan)
	logging.WithID("BA-OPERATOR-VERIFIER-13").Info(util.StatValuesToString(orphan))

	down, err := verifyOSDDown(dataset["OSD_UP"].Queue, dataset["OSDInQuorum"].Queue, length)
	*struc = append(*struc, down)
	logging.WithID("BA-OPERATOR-VERIFIER-17").Info(util.StatValuesToString(down))

	stale, err := verifyPG(dataset["PG_Stale"].Queue, length, "PG_Stale")
	*struc = append(*struc, stale)
	logging.WithID("BA-OPERATOR-VERIFIER-18").Info(util.StatValuesToString(stale))

	degraded, err := verifyPG(dataset["PG_Degraded"].Queue, length, "PG_Degraded")
	*struc = append(*struc, degraded)
	logging.WithID("BA-OPERATOR-VERIFIER-19").Info(util.StatValuesToString(degraded))

	undersized, err := verifyPG(dataset["PG_Undersized"].Queue, length, "PG_Undersized")
	*struc = append(*struc, undersized)
	logging.WithID("BA-OPERATOR-VERIFIER-19").Info(util.StatValuesToString(undersized))

	// tpRead, tpReadStatus, tpReadDev, tpWarning, err := verifyTPRead(dataset["TPread"].Queue, length)
	// data[7] = GetStatValues("throughput read", tpRead, tpReadStatus, tpReadDev, tpWarning)

	// tpWrite, tpWriteStatus, tpWriteDev, tpWarning, err := verifyTPWrite(dataset["TPread"].Queue, length)
	// data[10] = GetStatValues("throughput write", tpWrite, tpWriteStatus, tpWriteDev, tpWarning)

	return err
}

// VerfiyInfrastructureStatus analyse infrastructure status
func VerfiyInfrastructureStatus(struc *[]model.StatValues, dataset map[string]queue.Dataset, length int) (int, error) {
	yellow := 0
	red := 0

	cpu, err := verifyCPUUsage(dataset["PercUsedCPU"].Queue, length)
	logging.WithID("BA-OPERATOR-VERIFIER-03").Info(util.StatValuesToString(cpu))
	*struc = append(*struc, cpu)
	if cpu.ValueStatus == model.DEGRADED {
		yellow += 3
	} else if cpu.ValueStatus == model.ERROR {
		red += 3
	}

	cores, err := verifyCPUCoresUsage(dataset["CPUCoresUsed"].Queue, length)
	logging.WithID("BA-OPERATOR-VERIFIER-04").Info(util.StatValuesToString(cores))
	*struc = append(*struc, cores)
	if cores.ValueStatus == model.DEGRADED {
		yellow += 3
	} else if cores.ValueStatus == model.ERROR {
		red += 3
	}

	memory, err := verifyMemUsage(dataset["UsePercOfMem"].Queue, length)
	logging.WithID("BA-OPERATOR-VERIFIER-05").Info(util.StatValuesToString(memory))
	*struc = append(*struc, memory)
	if memory.ValueStatus == model.DEGRADED {
		yellow += 3
	} else if memory.ValueStatus == model.ERROR {
		red += 3
	}

	network, err := verifyNetworkUsage(dataset["networkTransmit"].Queue, length)
	logging.WithID("BA-OPERATOR-VERIFIER-06").Info(util.StatValuesToString(network))
	*struc = append(*struc, network)
	if network.ValueStatus == model.DEGRADED {
		yellow += 2
	} else if network.ValueStatus == model.ERROR {
		red += 2
	}

	capacity, err := verifyCapUsage(dataset["Av_capacity"].Queue, length)
	logging.WithID("BA-OPERATOR-VERIFIER-07").Info(util.StatValuesToString(capacity))
	*struc = append(*struc, capacity)
	if capacity.ValueStatus == model.DEGRADED {
		yellow++
	} else if capacity.ValueStatus == model.ERROR {
		red++
	}

	daysRemainingCap := predictDaysToCapacitiyLimit(dataset["Av_capacity"].Queue, length)
	logging.WithID("BA-OPERATOR-VERIFIER-15").Info("Predicted Day until Memory: " + util.StatusToStr(daysRemainingCap))

	if red >= 1 {
		return model.ERROR, err
	} else if yellow >= 2 {
		return model.DEGRADED, err
	} else {
		return model.HEALTHY, err
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

func verifyTPRead(queue *queue.MetricQueue, length int) (model.StatValues, error) {
	commit := queue.GetNNewestTupel(length)
	if len(commit) == 0 {
		return util.GetStatValuesEmpty("TPread"), nil
	}
	mean := stats.Mean(commit, length)
	max := stats.Max(commit, length)
	_, perc := stats.GetQuantiles(commit, config)
	fmt.Println(commit, mean, max, perc)

	deviation := stats.Deviation(commit, length)
	deviation += mean

	status := model.HEALTHY
	devStatus := model.HEALTHY
	limitYellow := 10.00
	limitRed := 50.00

	if deviation > limitRed {
		devStatus = model.ERROR
	} else if deviation >= limitYellow && deviation <= limitRed {
		devStatus = model.DEGRADED
	}

	if perc[0.90] > limitRed {
		devStatus = model.ERROR
	} else if perc[0.90] >= limitYellow && perc[0.90] <= limitRed {
		devStatus = model.DEGRADED
	}
	return util.GetStatValuesDev("TPread", perc[0.90], status, deviation, devStatus), nil

}

func verifyTPWrite(queue *queue.MetricQueue, length int) (model.StatValues, error) {
	commit := queue.GetNNewestTupel(length)
	if len(commit) == 0 {
		return util.GetStatValuesEmpty("TPwrite"), nil
	}
	mean := stats.Mean(commit, length)
	max := stats.Max(commit, length)
	_, perc := stats.GetQuantiles(commit, config)
	fmt.Println(commit, mean, max, perc)

	deviation := stats.Deviation(commit, length)
	deviation += mean

	status := model.HEALTHY
	devStatus := model.HEALTHY
	limitYellow := 10.00
	limitRed := 50.00

	if deviation > limitRed {
		devStatus = model.ERROR
	} else if deviation >= limitYellow && deviation <= limitRed {
		devStatus = model.DEGRADED
	}

	if perc[0.90] > limitRed {
		devStatus = model.ERROR
	} else if perc[0.90] >= limitYellow && perc[0.90] <= limitRed {
		devStatus = model.DEGRADED
	}
	return util.GetStatValuesDev("commit", perc[0.90], status, deviation, devStatus), nil

}
