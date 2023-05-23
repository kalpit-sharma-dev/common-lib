package model

import "github.com/gocql/gocql"

//go:generate repo -package=model -out=./cat_repository_gen.go -table=cats -keys="ID,Age" -viewTables=[cats_by_age:"Age,ID";cats_by_name:"Name,Age,ID"] "entity=cat entities=cats"

// Cat holds information about our pet
type Cat struct {
	ID          gocql.UUID
	Name        string
	Age         int
	WeightValue int `db:"Weight"`
	Owners      []string
}

// AcquireID acquire ID for cat if it not present
func (c *Cat) AcquireID() error {
	if c.ID != (gocql.UUID{}) {
		return nil
	}
	id, err := gocql.RandomUUID()
	if err != nil {
		return err
	}
	c.ID = id
	return nil
}
