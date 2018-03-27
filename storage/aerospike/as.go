package aerospike

import (
	"git.workshop21.ch/ewa/common/go/abraxas/logging"
	"git.workshop21.ch/ewa/common/go/abraxas/storage/aerospike"
	cmncfg "git.workshop21.ch/ewa/common/go/vrsg/configuration"
	"git.workshop21.ch/workshop21/ba/operator/configuration"
	"git.workshop21.ch/workshop21/ba/operator/model"
	as "github.com/aerospike/aerospike-client-go"
	as_types "github.com/aerospike/aerospike-client-go/types"
)

type ASStorage struct {
	client   *aerospike.Client
	RepoData aerospike.Repository
}

func NewASStorage(config *configuration.Config) (*ASStorage, error) {
	asClient, err := aerospike.NewClient(config.AerospikeConfig)
	return &ASStorage{client: asClient}, err
}
func (s *ASStorage) CreateData(data *model.Data) error {
	dataModel := DataModel{}
	err := s.RepoData.GetByID(data.Timestamp, &dataModel)
	if err != nil {
		if isAEROType(err, as_types.KEY_NOT_FOUND_ERROR) {
			_, err = s.RepoData.CreateObject(data)
			if err != nil {
				return err
			}
		} else {
			logging.WithError("BA-OPERATOR-ksjn42", err).Error("Unable to create data")
			return err
		}

	}

	return err
}

func isAEROType(err error, t as_types.ResultCode) bool {
	if AERO, ok := err.(as_types.AerospikeError); ok {
		return AERO.ResultCode() == t
	}
	return false
}

type DataModel struct {
	ID     int
	Values map[string]float64
}

//GetVrsgClientPolicy returns the default client policy for our applications based on the aerospike default policy
func GetVrsgClientPolicy() *as.ClientPolicy {
	pol := as.NewClientPolicy()
	pol.RequestProleReplicas = true
	return pol
}

// GetAeroHosts creates an array of aerospike hosts from config
func GetAeroHosts(config *cmncfg.AerospikeConfig) []*as.Host {
	hosts := make([]*as.Host, 0)
	for _, host := range config.AerospikeHosts {
		hosts = append(hosts, as.NewHost(host, config.AerospikePort))
	}
	return hosts
}
