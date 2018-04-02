package storage

import (
	"fmt"
	"log"
	"strconv"
	"time"

	"git.workshop21.ch/workshop21/ba/operator/configuration"
	as "github.com/aerospike/aerospike-client-go"
)

type ASStorage struct {
	Client  *as.Client
	Policy  *as.BasePolicy
	WPolicy *as.WritePolicy
}

type MonitorRecord struct {
	IOPS_write                 MonitorBin
	IOPS_read                  MonitorBin
	Monitors_quorum            MonitorBin
	Available_capacity         MonitorBin
	AverageMonitorLatency      MonitorBin
	Average_OSD_apply_latency  MonitorBin
	Average_OSD_commit_latency MonitorBin
	Throughput_write           MonitorBin
	Throughput_read            MonitorBin
	CEPH_health                MonitorBin
	OSD_Orphans                MonitorBin
	Used_percent_of_cores      MonitorBin
	Used_percent_of_memory     MonitorBin
	network_usage              MonitorBin
}

type MonitorBin struct {
	Bin   string
	Value float64
}

func CreateClient(config *configuration.Config) (*ASStorage, error) {
	log.Println(config.AerospikeHost, config.AerospikePort)
	asClient, err := as.NewClient("localhost", 3000)
	if err != nil {
		log.Println(err.Error())
	}
	policy := as.NewPolicy()
	policy.Timeout = 100 * time.Millisecond
	wPolicy := as.NewWritePolicy(0, 0)
	wPolicy.Timeout = 100 * time.Millisecond
	return &ASStorage{Client: asClient, Policy: policy, WPolicy: wPolicy}, err
}

func (s ASStorage) WriteBin(key int, value float64, bin string) error {
	asKey, err := s.GetKey(strconv.Itoa(key))
	if err != nil {
		log.Println(err.Error())
		return err
	}
	// Write multiple values.
	asBin := as.NewBin(bin, value)
	return s.Client.PutBins(s.WPolicy, asKey, asBin)
}

// func (s ASStorage) WriteRecord(key int, bins map[string]float64) error {
// 	asKey, err := s.GetKey(strconv.Itoa(key))
// 	if err != nil {
// 		log.Println(err.Error())
// 		return err
// 	}
// 	var asBins []*as.Bin
// 	// Write multiple values.
// 	asBin := as.NewBin("Timestamp", key)
// 	asBins = append(asBins, asBin)
// 	for binName, binValue := range bins {
// 		newBin := as.NewBin(binName, binValue)
// 		asBins := append(asBins, newBin)
// 	}
// 	return s.Client.PutBins(s.WPolicy, asKey, asBins)
// }

func (s ASStorage) GetKey(val string) (*as.Key, error) {
	key, err := as.NewKey("test", "set",
		val)
	if err != nil {
		log.Println(err)
	}
	return key, err
}

func (s ASStorage) ReadRecord(key int) (*as.Record, error) {
	asKey, err := s.GetKey(strconv.Itoa(key))
	if err != nil {
		return nil, err
	}
	exists, err := s.Client.Exists(s.Policy, asKey)
	if err != nil {
		return nil, err
	}
	if exists {
		record, err := s.Client.Get(s.Policy, asKey)
		if err != nil {
			return nil, err
		}

		fmt.Println(record)
		return record, err

	}
	return nil, err
}
