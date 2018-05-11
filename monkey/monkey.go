package monkey

import (
	"math/rand"
	"time"

	"git.workshop21.ch/go/abraxas/logging"
	"git.workshop21.ch/workshop21/ba/operator/configuration"
	"git.workshop21.ch/workshop21/ba/operator/kubeclient"
	"git.workshop21.ch/workshop21/ba/operator/monitoring"
)

var kc *kubeclient.KubeClient

func DoTheMonkey(config *configuration.Config) {

	components := make([]string, 0)
	components = append(components,
		config.RookMonSelector,
		config.RookOSDSelector)
	logging.WithID("MONKEY-000").Info("monkey started going wild")
	for {
		time.Sleep(time.Duration(13) * time.Minute)

		if monitoring.VerifyClusterStatus() {
			s := rand.NewSource(time.Now().Unix())
			r := rand.New(s) // initialize local pseudorandom genergoator

			var err error
			kc, err = kubeclient.GetKubeClient(kc)
			if err != nil {
				logging.WithID("MONKEY-001").Error("\nnot able to get kubeclient " + err.Error())
				return
			}
			component := components[r.Intn(len(components))]
			logging.WithID("MONKEY-001-1").Info("monkey is shooting coconut against ", component)
			err = kc.KillOnePodOf(component)
			if err != nil {
				logging.WithID("MONKEY-002").Error("\nnot able to kill a pod out of " + component + err.Error())
			}
		} else {
			logging.WithID("MONKEY-003").Error("not able to kill pod. Cluster is not ready!")
		}

	}

}
