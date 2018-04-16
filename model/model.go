package model

import "github.com/hishboy/gocommons/lang"

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

type Dataset struct {
	Set   map[int]float64
	Queue *lang.Queue
}
