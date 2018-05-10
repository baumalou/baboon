package monkey

import (
	"math/rand"
	"time"

	"git.workshop21.ch/go/abraxas/logging"
	"git.workshop21.ch/workshop21/ba/operator/kubeclient"
	"git.workshop21.ch/workshop21/ba/operator/monitoring"
)

var kc *kubeclient.KubeClient

func DoTheMonkey() {
	reasons := make([]string, 0)
	reasons = append(reasons,
		"mon",
		"osd")
	logging.WithID("MONKEY-000").Info("monkey started going wild")
	for {
		time.Sleep(time.Duration(random(1, 10)) * time.Minute)
		if monitoring.VerifyClusterStatus() {
			s := rand.NewSource(time.Now().Unix())
			r := rand.New(s) // initialize local pseudorandom generator

			var err error
			kc, err = kubeclient.GetKubeClient(kc)
			if err != nil {
				logging.WithID("MONKEY-001").Error("\nnot able to get kubeclient " + err.Error())
				return
			}
			component := reasons[r.Intn(len(reasons))]
			err = kc.KillOnePodOf(component)
			if err != nil {
				logging.WithID("MONKEY-002").Error("\nnot able to kill a pod out of " + component + err.Error())
			}
		} else {
			logging.WithID("MONKEY-003").Error("\nnot able to kill pod. Cluster is not ready!")
		}

	}

}
