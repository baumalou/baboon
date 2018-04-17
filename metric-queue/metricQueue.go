package queue

type Dataset struct {
	Name  string
	Set   []MetricTupel
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

func (mq *MetricQueue) removeOldestItem() MetricTupel {
	if len(mq.Dataset) > 0 {
		mq.Dataset = mq.Dataset[1:]
		return mq.Dataset[0]
	}
	return MetricTupel{}

}

func (mq *MetricQueue) AddMonitoringTupelSliceToDataset(tupelArray []MetricTupel) {
	for _, tupel := range tupelArray {
		mq.Push(tupel)
	}
}

func (mq *MetricQueue) Push(tupel MetricTupel) {
	mq.Dataset = append(mq.Dataset, tupel)
}

func (mq *MetricQueue) Pop() MetricTupel {
	return mq.removeOldestItem()
}
