package verifier

import (
	queue "git.workshop21.ch/workshop21/ba/operator/metric-queue"
	stats "git.workshop21.ch/workshop21/ba/operator/statistics"
)

type Status int

const (
	HEALTHY Status = 1 + iota
	DEGRADED
	ERROR
)

// VerifyClusterStatus func cluster
func VerifyClusterStatus(dataset map[string]queue.Dataset) Status {
	length := 4

	iops := verifyIOPS(dataset["IOPS_write"].Queue, dataset["IOPS_read"].Queue, length)
	mon := verifyMonitorCounts(dataset["Mon_quorum"].Queue.Dataset, length)
	commit := verifyOSDCommitLatency(dataset["AvOSDcommlat"].Queue.Dataset, length)
	apply := verifyOSDApplyLatency(dataset["AvOSDappllat"].Queue.Dataset, length)
	health := verifyCephHealth(dataset["CEPH_health"].Queue.Dataset, length)
	orphan := verifyOSDOrphan(dataset["OSDInQuorum"].Queue, dataset["OSD_UP"].Queue, length)

	infra := VerfiyInfrastructureStatus(dataset, length)

	if iops == ERROR || mon == ERROR || commit == ERROR || apply == ERROR || health == ERROR || orphan == ERROR || infra == ERROR {
		return ERROR
	} else if iops == DEGRADED || mon == DEGRADED || commit == DEGRADED || apply == DEGRADED || health == DEGRADED || orphan == DEGRADED || infra == DEGRADED {
		return DEGRADED
	} else {
		return HEALTHY
	}

}

// VerfiyInfrastructureStatus func infrastructure
func VerfiyInfrastructureStatus(dataset map[string]queue.Dataset, length int) Status {
	yellow := 0
	red := 0

	cpu := verifyCPUUsage(dataset["PercUsedCPU"].Queue.Dataset, length)
	if cpu == DEGRADED {
		yellow += 3
	} else if cpu == ERROR {
		red += 3
	}
	cores := verifyCPUCoresUsage(dataset["CPUCoresUsed"].Queue.Dataset, length)
	if cores == DEGRADED {
		yellow += 3
	} else if cores == ERROR {
		red += 3
	}
	mem := verifyMemUsage(dataset["UsePercOfMem"].Queue.Dataset, length)
	if mem == DEGRADED {
		yellow += 3
	} else if mem == ERROR {
		red += 3
	}

	net := verifyNetworkUsage(dataset["networkReceive"].Queue, dataset["networkSend"].Queue, length)
	if net == DEGRADED {
		yellow += 2
	} else if net == ERROR {
		red += 2
	}

	cap := verifyCapUsage(dataset["Av_capacity"].Queue.Dataset, length)
	if cap == DEGRADED {
		yellow++
	} else if cap == ERROR {
		red++
	}

	if red >= 1 {
		return ERROR
	} else if yellow >= 2 {
		return DEGRADED
	} else {
		return HEALTHY
	}
}

func verifyIOPS(write *queue.MetricQueue, read *queue.MetricQueue, length int) Status {
	writeDS := write.GetNNewestTupel(length)
	readDS := read.GetNNewestTupel(length)
	data := make([]queue.MetricTupel, length)
	for i := 0; i < length; i++ {
		data[i].Timestamp = writeDS[i].Timestamp
		data[i].Value = writeDS[i].Value + readDS[i].Value
	}
	result := stats.Mean(data, length)
	max := stats.Max(data, length)
	min := stats.Min(data, length)
	deviation := stats.Deviation(data, length)

	if (max-deviation) > result || (min+deviation) < result {
		return HEALTHY
	} else {
		if result < 6000 {
			return HEALTHY
		} else if result > 6000 && result < 14000 {
			return DEGRADED
		} else {
			return ERROR
		}
	}
}

func verifyMonitorCounts(data []queue.MetricTupel, length int) Status {
	min := stats.Min(data, length)
	if min < 2 {
		return ERROR
	} else if min < 3 {
		return DEGRADED
	} else {
		return HEALTHY
	}
}

func verifyOSDCommitLatency(commit []queue.MetricTupel, length int) Status {

	max := stats.Max(commit, length)

	if max > 50 {
		return ERROR
	} else if max >= 10 && max <= 50 {
		return DEGRADED
	} else {
		return HEALTHY
	}
}
func verifyOSDApplyLatency(apply []queue.MetricTupel, length int) Status {
	max := stats.Max(apply, length)

	if max > 50 {
		return ERROR
	} else if max >= 10 && max <= 50 {
		return DEGRADED
	} else {
		return HEALTHY
	}
}

func verifyCephHealth(health []queue.MetricTupel, length int) Status {
	max := stats.Max(health, length)

	if max == 2 {
		return ERROR
	} else if max == 1 {
		return DEGRADED
	} else {
		return HEALTHY
	}
}

func verifyCPUUsage(usage []queue.MetricTupel, length int) Status {

	result := stats.Mean(usage, length)
	//max := stats.Max(usage, length)
	//min := stats.Min(usage, length)
	//deviation := stats.Deviation(usage, length)

	if result > 85 {
		return ERROR
	} else if result >= 10 && result <= 85 {
		return DEGRADED
	} else {
		return HEALTHY
	}
}

func verifyCPUCoresUsage(usage []queue.MetricTupel, length int) Status {

	result := stats.Mean(usage, length)
	//max := stats.Max(usage, length)
	//min := stats.Min(usage, length)
	//deviation := stats.Deviation(usage, length)

	if result > 85 {
		return ERROR
	} else if result >= 10 && result <= 85 {
		return DEGRADED
	} else {
		return HEALTHY
	}
}

func verifyMemUsage(usage []queue.MetricTupel, length int) Status {

	result := stats.Mean(usage, length)
	//max := stats.Max(usage, length)
	//min := stats.Min(usage, length)
	//deviation := stats.Deviation(usage, length)

	if result > 80 {
		return ERROR
	} else if result >= 10 && result <= 80 {
		return DEGRADED
	} else {
		return HEALTHY
	}
}
func verifyNetworkUsage(rec *queue.MetricQueue, send *queue.MetricQueue, length int) Status {
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
		return ERROR
	} else if result >= 10 && result <= 80 {
		return DEGRADED
	} else {
		return HEALTHY
	}
}

func verifyOSDOrphan(in *queue.MetricQueue, up *queue.MetricQueue, length int) Status {
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
		return ERROR
	} else if max == 1 {
		return DEGRADED
	} else {
		return HEALTHY
	}
}
func verifyCapUsage(usage []queue.MetricTupel, length int) Status {

	result := stats.Mean(usage, length)
	//max := stats.Max(usage, length)
	//min := stats.Min(usage, length)
	//deviation := stats.Deviation(usage, length)

	if result > 80 {
		return ERROR
	} else if result >= 10 && result <= 80 {
		return DEGRADED
	} else {
		return HEALTHY
	}
}
