package model

import "time"

//go:generate repo -package=model -out=./dog_repository_gen.go -table=dogs -type=string -keys="Name" "entity=dog entities=dogs"

// Dog holds information about our pet
type Dog struct {
	Name        string
	Age         int
	ValueWeight int `db:"Weight"`
	Owners      []string
	DateOfBirth time.Time
}

// AcquireID acquire Name for dog if it not present
func (d *Dog) AcquireID() error {
	if d.Name != "" {
		return nil
	}
	// add some code to set really unique name
	d.Name = "Pluto"
	return nil
}
