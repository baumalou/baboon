package model

type GrafanaResult struct {
	Status string `json:"status"`
	Data   struct {
		ResultType string `json:"resultType"`
		Result     []struct {
			Metric struct {
			} `json:"metric"`
			Values [][]interface{} `json:"values"`
		} `json:"result"`
	} `json:"data"`
}

type Data struct {
	Bin       string
	Timestamp int
	Value     float64
}

type StatValues struct {
	Name          string
	Value         float64
	ValueStatus   int
	DevValue      float64
	DevStatus     int
	PercentileVal float64
}
