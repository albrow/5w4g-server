package models

import (
	"code.google.com/p/go.crypto/bcrypt"
	"github.com/albrow/5w4g-server/config"
	"github.com/albrow/zoom"
)

func Init() {

	// Initialize zoom
	zoom.Init(&zoom.Configuration{
		Network:  config.Db.Network,
		Address:  config.Db.Address,
		Database: config.Db.Database,
	})

	// Register all models
	if err := zoom.Register(&AdminUser{}); err != nil {
		panic(err)
	}

	// Create a default admin user if needed
	if err := CreateDefaultAdminUser(); err != nil {
		panic(err)
	}
}

// Identifier has an Id field and satisfies zoom.Model.
// The only difference is that it has a tag to convert the
// Id to lowercase in json, which is the format ember.js expects.
type Identifier struct {
	Id string `json:"id"`
}

func (i *Identifier) GetId() string {
	return i.Id
}

func (i *Identifier) SetId(id string) {
	i.Id = id
}

func CreateDefaultAdminUser() error {
	if config.Env == "test" {
		return nil
	}
	admins := []*AdminUser{}
	if err := zoom.NewQuery("AdminUser").Scan(&admins); err != nil {
		return err
	}
	if len(admins) == 0 {
		// No admins exist, so we should create one with a default email and password
		hash, err := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.DefaultCost)
		if err != nil {
			return err
		}
		defaultUser := &AdminUser{
			Email:          "admin@5w4g.com",
			HashedPassword: string(hash),
		}
		if err := zoom.Save(defaultUser); err != nil {
			return err
		}
	}
	return nil
}
