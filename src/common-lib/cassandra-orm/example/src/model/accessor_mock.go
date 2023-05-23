package model

import (
	"github.com/gocql/gocql"
	db "gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/cassandra-orm"
)

var (
	tomID     = gocql.TimeUUID()
	plutoName string
)

// MockCats mock and fill cats
func MockCats() {
	catMock := db.NewAccessorMock(catBaseTableName, catKeyColumns, &Cat{}, catViewTables)
	catMock.Register(&CatObserver{})
	Cats().base = catMock

	cats := []*Cat{
		{
			ID:          tomID,
			Name:        "Tom",
			Age:         5,
			WeightValue: 8,
			Owners:      []string{"Molly"},
		},
		{ // ID for this field will be set via SetID automatically
			Name:        "Jerry",
			Age:         3,
			WeightValue: 5,
			Owners:      []string{"Bob", "Mary"},
		},
	}
	for _, cat := range cats {
		err := Cats().Add(cat)
		if err != nil {
			panic(err)
		}
	}
}

// MockDogs mock and fill dogs
func MockDogs() {
	dogMock := db.NewAccessorMock(dogBaseTableName, dogKeyColumns, &Dog{}, nil)
	Dogs().base = dogMock

	dogs := []*Dog{
		{
			Age:         5,
			ValueWeight: 8,
			Owners:      []string{"Polly"},
		},
		{
			Name:        "Charlie",
			Age:         3,
			ValueWeight: 5,
			Owners:      []string{"Mr.Brown", "Mrs.Brown"},
		},
	}
	for _, dog := range dogs {
		err := Dogs().Add(dog)
		if err != nil {
			panic(err)
		}
	}
	plutoName = dogs[0].Name
}
