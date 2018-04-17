package verifier

import queue "git.workshop21.ch/workshop21/ba/operator/metric-queue"

type Status int

const (
	HEALTHY Status = 1 + iota
	DEGRADED
	ERROR
)

func VerifyClusterStatus(dataset map[string]queue.Dataset) (Status, error) {
	return HEALTHY, nil
}
