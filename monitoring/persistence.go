package monitoring

import (
	"git.workshop21.ch/go/abraxas/logging"
	"git.workshop21.ch/workshop21/ba/operator/storage"
)

func storeDataset(dataSet map[int]float64, keys []int, binName string, asStorage *storage.ASStorage, set string) error {
	index := 0
	for _, value := range dataSet {
		if len(keys) <= index {
			logging.WithID("BA-OPERATOR-storeDataset-01").Println(len(keys), index, "index exceeds")
			return nil
		}
		err := asStorage.WriteBin(keys[index], value, binName, set)
		index++
		if err != nil {
			logging.WithID("BA-OPERATOR-storeDataset-02").Println(err.Error(), err)
		}
	}
	return nil
}
