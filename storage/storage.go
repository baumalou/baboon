package storage

// import (
// 	"encoding/json"
// 	"io/ioutil"
// 	"os"
// 	"path"

// 	"git.workshop21.ch/ewa/common/go/abraxas/logging"
// 	"git.workshop21.ch/ewa/common/go/abraxas/storage"
// 	"git.workshop21.ch/ewa/common/go/abraxas/storage/filter"
// 	"git.workshop21.ch/ewa/common/go/abraxas/storage/idSupplier"
// 	"git.workshop21.ch/ewa/common/go/abraxas/storage/migration"
// )

// type Storage interface {
// 	Health() error
// 	MigratorRepository() storage.Repository
// 	UserStorage
// 	LoginStorage
// }

// type UserStorage interface {
// 	GetUserList(...filter.Filter) (*id_pb.Users, error)
// 	GetUserByID(id interface{}) (*id_pb.User, error)
// 	GetUserByLoginID(id interface{}) (*id_pb.User, error)
// 	UpdateUser(*id_pb.User) (*id_pb.User, error)
// 	CreateUser(*id_pb.User) (*id_pb.User, error)
// 	DeleteUser(id interface{}) error
// 	GetActiveUserFilter() filter.PredicateFunc
// }

// type LoginStorage interface {
// 	GetLoginByID(interface{}) (*id_pb.Login, error)
// 	GetLoginByIdmID(string) (*id_pb.Login, error)
// 	GetLoginByUsername(string) (*id_pb.Login, error)
// 	UpdateLogin(*id_pb.Login) (*id_pb.Login, error)
// 	CreateLogin(*id_pb.Login) (*id_pb.Login, error)
// }

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
