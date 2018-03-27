package storage

import (
	"git.workshop21.ch/ewa/common/go/abraxas/storage"
	"git.workshop21.ch/workshop21/ba/operator/model"
)

type Storage interface {
	Health() error
	MigratorRepository() storage.Repository
}

type DataStorage interface {
	CreateData(*model.Data) error
}

// func NewID() uint64 {
// 	id, err := idSupplier.SonyflakeIDSupplier.NewID(nil)
// 	if err != nil {
// 		logging.WithError("AERO-3be3b758", err).Fatal("error generating sonyflake id")
// 	}
// 	return id.(uint64)
// }

// func Migrate(s Storage) Storage {
// 	if err := migration.ExecuteRegisteredMigartions(s.MigratorRepository()); err != nil {
// 		logging.WithError("STOR-791d7ce9", err).Fatal("error migrating db")
// 	}
// 	return s
// }

// // ReadDefaults unmarshals default objects to go structs
// func ReadDefaults(filename string, i interface{}) error {
// 	pathToDefaults := path.Join(pathForResources(), filename)

// 	logging.WithID("IDENTITY-o24ftkh").WithField("Path", pathToDefaults).Info("Reading defaultsfile")

// 	jsonData, err := ioutil.ReadFile(pathToDefaults)
// 	if err != nil {
// 		return err
// 	}
// 	return json.Unmarshal(jsonData, i)
// }

// //PathForResources path for resources (looks for env)
// func pathForResources() string {
// 	wd, _ := os.Getwd()
// 	return path.Join(wd, "storage", "resources")
// }
