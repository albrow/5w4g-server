package models

import (
	"code.google.com/p/go.crypto/bcrypt"
	"github.com/albrow/5w4g-server/config"
	"github.com/albrow/zoom"
	"sync"
)

var once = sync.Once{}

func Init() {
	once.Do(func() {
		// Initialize zoom
		zoom.Init(&zoom.Configuration{
			Network:  config.Db.Network,
			Address:  config.Db.Address,
			Database: config.Db.Database,
		})

		// Register all models
		models := []zoom.Model{&AdminUser{}, &Item{}, &OrderItem{}, &Order{}}
		for _, m := range models {
			if err := zoom.Register(m); err != nil {
				panic(err)
			}
		}

		// If we're in test environment, flush the database on startup
		if config.Env == "test" {
			conn := zoom.GetConn()
			if _, err := conn.Do("FLUSHDB"); err != nil {
				panic(err)
			}
		}

		// Create a default admin user if needed
		if err := CreateDefaultAdminUser(); err != nil {
			panic(err)
		}
	})
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
	if num, err := zoom.NewQuery("AdminUser").Count(); err != nil {
		return err
	} else if num == 0 {
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
