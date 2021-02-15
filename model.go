package main

import (
	"os/user"
	"time"

	"github.com/go-pg/pg/v10"
	"github.com/go-pg/pg/v10/orm"
)

// PingURL - Model of the URL to be pinged
type PingURL struct {
	ID          int64
	Name        string     `sql:",unique,notnull"`
	URL         string     `sql:",notnull"`
	Status      *user.User `sql:",notnull"`
	Count       int64      `sql:",notnull"`
	PassedCount int64      `sql:",notnull"`
	LastChecked time.Time  `sql:",notnull"`
	CreatedDate time.Time  `sql:",notnull"`
}

//CreateSchema for Stories
func CreateSchema(db *pg.DB) error {
	models := []interface{}{
		(*PingURL)(nil),
		// For more models here.
	}

	for _, model := range models {
		err := db.Model(model).CreateTable(&orm.CreateTableOptions{
			Temp: true,
		})
		if err != nil {
			return err
		}
	}
	return nil
}
