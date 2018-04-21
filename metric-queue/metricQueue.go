package queue

import (
	"sort"
)

type Dataset struct {
	Name  string
	Queue *MetricQueue
}
type MetricTupel struct {
	Timestamp int
	Value     float64
}

type MetricQueue struct {
	Dataset []MetricTupel
}

func NewMetricQueue() *MetricQueue {
	return &MetricQueue{Dataset: nil}
}

func (mq *MetricQueue) sort() {
	sort.Sort(mq)
}

func (mq *MetricQueue) Swap(i, j int) {
	temp := mq.Dataset[i]
	mq.Dataset[i] = mq.Dataset[j]
	mq.Dataset[j] = temp
}

func (mq *MetricQueue) Less(i, j int) bool {
	return mq.Dataset[i].Timestamp < mq.Dataset[j].Timestamp
}

func (mq *MetricQueue) Len() int {
	return len(mq.Dataset)
}

func (mq *MetricQueue) removeOldestItem() MetricTupel {
	if len(mq.Dataset) > 0 {
		metricTupeToReturn := mq.Dataset[0]
		mq.Dataset = mq.Dataset[1:]
		return metricTupeToReturn
	}
	return MetricTupel{}

}

func (mq *MetricQueue) AddMonitoringTupelSliceToDataset(tupelArray []MetricTupel) {
	for _, tupel := range tupelArray {
		mq.Push(tupel)
		mq.Pop()
	}
}

func (mq *MetricQueue) InsertMonitoringTupelInQueue(tupelArray []MetricTupel) {
	for _, tupel := range tupelArray {
		mq.Push(tupel)
	}
}

func (mq *MetricQueue) GetNNewestTupel(n int) []MetricTupel {
	length := len(mq.Dataset)
	data := make([]MetricTupel, n)
	if length > 0 {
		mq.Dataset = mq.Dataset[1:]
		for i := 0; i < n; i++ {
			data[n-i] = mq.Dataset[length-i]
		}
	}
	return data
}

func (mq *MetricQueue) Push(tupel MetricTupel) {
	mq.Dataset = append(mq.Dataset, tupel)
}

func (mq *MetricQueue) Pop() MetricTupel {
	return mq.removeOldestItem()
}
