package model

import (
	"github.com/jmoiron/sqlx/reflectx"
	"github.com/scylladb/gocqlx"
)

func init() {
	// to control mapping between field name at you model (struct) and db field name use one of:
	// reflectx.NewMapper, reflectx.NewMapperFunc or reflectx.NewMapperTagFunc
	// reflectx.NewMapper - set mapping as-is - cat.PetName will store at db at field "PetName"
	// reflectx.NewMapperFunc("db", snakeCase) - cat.PetName will store at db at field "pet_name"
	// in all cases you can use `db:"-"` to not store field at db
	gocqlx.DefaultMapper = reflectx.NewMapper("db")
	//gocqlx.DefaultMapper = reflectx.NewMapperFunc("db", snakeCase)
}
