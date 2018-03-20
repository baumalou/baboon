package configuration

import (
	"log"
	"os"
	"path/filepath"
	"runtime"

	cmnconfig "git.workshop21.ch/ewa/common/go/vrsg/configuration"

	"github.com/BurntSushi/toml"
)

const (
	ASNamespace = "iks"
)

type Config struct {
	OSDS_UP_Endpoint      string
	MonitoringHost        string
	ServiceConfig         cmnconfig.ServiceConfig
	BearerToken           string
	AVG_OSD_APPLY_LATENCY string
	SamplingStepSize      string
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
		return getAbsPath() + "/config" + env + ".toml"
	}
	return "./configuration/config.toml"
}

func ReadConfig(config *Config) (*Config, error) {
	if config == nil {
		if _, err := toml.DecodeFile(PathForConfig(), &config); err != nil {
			log.Println("Config could not be decoded: ", err)
			return &Config{}, err
		}
	}
	return config, nil
}
