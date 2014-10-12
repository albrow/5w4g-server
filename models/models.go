package models

import (
	"github.com/albrow/5w4g/config"
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
