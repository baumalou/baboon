package monitoring

import "time"

func removeOldestItem(dataset map[int]float64) map[int]float64 {
	if len(dataset) < 1 {
		return dataset
	}
	smallestTS := int(time.Now().Unix())
	for key := range dataset {
		if key < smallestTS {
			smallestTS = key
		}
	}
	_, ok := dataset[smallestTS]
	if ok {
		delete(dataset, smallestTS)
	}
	return dataset
}
