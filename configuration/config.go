package configuration

import (
	"os"
	"path/filepath"
	"runtime"

	"git.workshop21.ch/go/abraxas/logging"
	cmn_as_conf "git.workshop21.ch/go/abraxas/storage/aerospike"

	"github.com/BurntSushi/toml"
)

const (
	ASNamespace = "ba"
)

type Config struct {
	OSDS_UP_Endpoint      string
	MonitoringHost        string
	BearerToken           string
	AVG_OSD_APPLY_LATENCY string
	SamplingStepSize      string
	AerospikeConfig       *cmn_as_conf.Config
	AerospikePort         int
	AerospikeHost         string
	Endpoints             map[string]Endpoint
	AerospikeNamespace    string
	Percentiles           []float64
	WebPort               string
	WebPath               string
	SampleInterval        int
	RookOSDSelector       string
	RookMonSelector       string
	RookNamespace         string
}
type Endpoint struct {
	Name string
	Path string
}

func (c *Config) GetAerospikeNamespace() string {
	namespace := ASNamespace
	if os.Getenv("TestMode") == "true" {
		namespace = namespace + "_test"
	}
	return namespace
}

func getAbsPath() string {
	_, b, _, _ := runtime.Caller(0)
	return filepath.Dir(b)
}

func PathForConfig() string {
	env := os.Getenv("ENV")
	if env != "" {
		logging.WithID("BA-OPERATOR-CONFIG-001").Println(getAbsPath() + "/config" + env + ".toml")
		return getAbsPath() + "/config" + env + ".toml"
	}
	return "./configuration/config.toml"
}

func ReadConfig(config *Config) (*Config, error) {
	if config == nil {
		if _, err := toml.DecodeFile(PathForConfig(), &config); err != nil {
			logging.WithID("BA-OPERATOR-CONFIG-002").Println("Config could not be decoded: ", err)
			return &Config{}, err
		}
	}
	return config, nil
}
