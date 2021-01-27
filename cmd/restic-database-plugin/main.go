package main

import (
	"log"
	"os"

	dbplugin "github.com/hashicorp/vault/sdk/database/v5"
)

func main() {
	apiClientMeta := &api.PluginAPIClientMeta{}
	flags := apiClientMeta.FlagSet()
	flags.Parse(os.Args[1:])

	err := Run(apiClientMeta.GetTLSConfig())
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}
}

func Run(apiTLSConfig *api.TLSConfig) error {
	dbType, err := New()
	if err != nil {
		return err
	}

	dbplugin.Serve(dbType.(dbplugin.Database), api.VaultPluginTLSProvider(apiTLSConfig))

	return nil
}

func New() (interface{}, error) {
	db, err := newDatabase()
	if err != nil {
		return nil, err
	}

	// This middleware isn't strictly required, but highly recommended to prevent accidentally exposing
	// values such as passwords in error messages. An example of this is included below
	db = dbplugin.NewDatabaseErrorSanitizerMiddleware(db, db.secretValues)
	return db, nil
}

type resticRepo struct {
	// Variables for the database
	password string
}

func newDatabase() (resticRepo, error) {
	// ...
	db := &resticRepo{
		// ...
	}
	return db, nil
}

func (db *resticRepo) secretValues() map[string]string {
	return map[string]string{
		db.password: "[password]",
	}
}
