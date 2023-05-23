package main

import (
	"fmt"
	"reflect"

	"github.com/gocql/gocql"
	"github.com/pkg/errors"
	db "gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/cassandra-orm"
	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/runtime/logger"

	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/cassandra-orm/example/src/model"
)

const emptyTable = "empty table"

func main() {
	log, err := logger.Create(logger.Config{
		FileName: "./platform-pets.log",
		LogLevel: logger.DEBUG,
	})
	if err != nil {
		panic(err)
	}

	cassandraHosts := []string{"localhost:9042"} // replace cassandra with your host
	cassandraKeyspace := "platform_pets_db"
	cassandraTimeout := "3s"

	err = db.Load(cassandraHosts, cassandraKeyspace, cassandraTimeout, log)
	if err != nil {
		panic(err)
	}

	checkCats()

	oneMoreCat := newCat("Rudolf")
	addCat(oneMoreCat)
	checkCats()

	oneMoreCat.WeightValue += 10
	updateCat(oneMoreCat)
	checkCats()

	deleteCat(oneMoreCat)
	checkCats()

	oneMoreDog := newDog("Bob")
	addDog(oneMoreDog)
	checkDog(oneMoreDog, true)

	oneMoreDog.ValueWeight += 10
	updateDog(oneMoreDog)
	checkDog(oneMoreDog, true)

	deleteDog(oneMoreDog)
	checkDog(oneMoreDog, false)

	fmt.Println("Done.")
}

func newCat(name string) *model.Cat {
	return &model.Cat{
		ID:          gocql.TimeUUID(),
		Name:        name,
		Age:         17,
		WeightValue: 230,
		Owners:      []string{"unknown"},
	}
}

func newDog(name string) *model.Dog {
	return &model.Dog{
		Name:        name,
		Age:         17,
		ValueWeight: 230,
		Owners:      []string{"unknown"},
	}
}

func checkCats() {
	baseCats := getAllBaseCats()
	displayCats(baseCats...)

	catsByName := getAllCatsByName()
	displayCats(catsByName...)

	catsByAge := getAllCatsByAge()
	displayCats(catsByAge...)

	if len(baseCats) != len(catsByAge) || len(baseCats) != len(catsByName) {
		fmt.Println("Different tables items size")
		panic("error")
	}
	err := compareCats(baseCats, catsByName)
	if err != nil {
		panic(errors.Wrap(err, "base and by name"))
	}

	err = compareCats(baseCats, catsByAge)
	if err != nil {
		panic(errors.Wrap(err, "base and by age"))
	}
}

func compareCats(one, two []*model.Cat) error {
	for _, firstCat := range one {
		found := false
		for _, secondCat := range two {
			if secondCat.ID == firstCat.ID {
				found = true
				eq := reflect.DeepEqual(firstCat, secondCat)
				if !eq {
					return errors.Errorf("Cats with ID %q are not equal: \nfirst %v\nsecond: %v", firstCat.ID, firstCat, secondCat)
				}
				break
			}
		}
		if !found {
			return errors.Errorf("not found cat with ID %q", firstCat.ID)
		}
	}

	return nil
}

func displayCats(cats ...*model.Cat) {
	for ind, cat := range cats {
		fmt.Println(ind, cat)
	}
	fmt.Println()
}

func getAllBaseCats() []*model.Cat {
	fmt.Println(" - Get all cats -")
	cats, err := model.Cats().All()
	if err != nil {
		if err == gocql.ErrNotFound {
			fmt.Println(emptyTable)
			return nil
		}
		panic(err)
	}
	return cats
}

func getAllCatsByName() []*model.Cat {
	fmt.Println(" - Get all cats by name -")
	cats, err := model.Cats().AllCatsByName()
	if err != nil {
		if err == gocql.ErrNotFound {
			fmt.Println(emptyTable)
			return nil
		}
		panic(err)
	}
	return cats
}

func getAllCatsByAge() []*model.Cat {
	fmt.Println(" - Get all cats by age -")
	cats, err := model.Cats().AllCatsByAge()
	if err != nil {
		if err == gocql.ErrNotFound {
			fmt.Println(emptyTable)
			return nil
		}
		panic(err)
	}
	return cats
}

func addCat(cat *model.Cat) {
	fmt.Printf(" - Add cat: %q with id: %q -\n", cat.Name, cat.ID)
	err := model.Cats().Add(cat)
	if err != nil {
		panic(err)
	}
}

func updateCat(cat *model.Cat) {
	fmt.Printf(" - Update cat: %q with id: %q -\n", cat.Name, cat.ID)
	err := model.Cats().Update(cat)
	if err != nil {
		panic(err)
	}
	fmt.Println()
}

func deleteCat(cat *model.Cat) {
	fmt.Printf(" - Delete cat: %q with id: %q -\n", cat.Name, cat.ID)
	err := model.Cats().Delete(cat)
	fmt.Println()
	if err != nil {
		panic(err)
	}
}

func addDog(dog *model.Dog) {
	fmt.Println(" - Add dog -")
	fmt.Println(dog)
	err := model.Dogs().Add(dog)
	if err != nil {
		panic(err)
	}
}

func updateDog(dog *model.Dog) {
	fmt.Println(" - Update dog -")
	fmt.Println(dog)
	err := model.Dogs().Update(dog)
	if err != nil {
		panic(err)
	}
}

func deleteDog(dog *model.Dog) {
	fmt.Println(" - Delete dog -")
	fmt.Println(dog)
	err := model.Dogs().Delete(dog)
	if err != nil {
		panic(err)
	}
}

func checkDog(dog *model.Dog, expectExisting bool) {
	dbDog, err := model.Dogs().GetByID(dog.Name)
	if err != nil {
		if err == gocql.ErrNotFound {
			if expectExisting {
				panic("expected row dog existing: %v")
			}
			fmt.Println(emptyTable)
			return
		}
		panic(err)
	}
	eq := reflect.DeepEqual(dbDog, dog)
	if !eq {
		panic(errors.Errorf("Dogs with name %q are not equal: \nfrom db %v\nsecond: %v", dog.Name, dbDog, dog))
	}
	fmt.Println("from db:")
	fmt.Println(dbDog)
	fmt.Println()
}
