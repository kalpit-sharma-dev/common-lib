package model

import (
	"testing"
	"time"

	"github.com/gocql/gocql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var expectedPluto = &Dog{
	Name:        "Pluto",
	Age:         5,
	ValueWeight: 8,
	Owners:      []string{"Polly"},
}

func TestDogRepository_All(t *testing.T) {
	dogs, err := Dogs().All()
	require.NoError(t, err)
	assert.Len(t, dogs, 2)
}

func TestDogRepository_GetByID(t *testing.T) {
	pluto, err := Dogs().GetByID(plutoName)
	require.NoError(t, err)
	assert.Equal(t, expectedPluto, pluto)
}

func TestDogRepository_Update(t *testing.T) {
	defer MockDogs()

	pluto, err := Dogs().GetByID(plutoName)
	require.NoError(t, err)

	pluto.Age += 1
	pluto.ValueWeight += 2
	pluto.Owners = append(pluto.Owners, "Sara")

	err = Dogs().Update(pluto)
	require.NoError(t, err)

	pluto2, err := Dogs().GetByID(plutoName)
	require.NoError(t, err)

	expectedDog := &Dog{
		Name:        plutoName,
		Age:         6,
		ValueWeight: 10,
		Owners:      []string{"Polly", "Sara"},
	}
	assert.Equal(t, expectedDog, pluto2)
}

func TestDogRepository_Delete(t *testing.T) {
	defer MockDogs()

	pluto, err := Dogs().GetByID(plutoName)
	require.NoError(t, err)

	err = Dogs().Delete(pluto)
	require.NoError(t, err)

	pluto, err = Dogs().GetByID(plutoName)
	assert.EqualError(t, err, gocql.ErrNotFound.Error())
	assert.Nil(t, pluto)
}

func TestDogRepository_Add(t *testing.T) {
	dogs := []*Dog{
		{
			Age:         5,
			ValueWeight: 8,
			Owners:      []string{"Polly"},
			DateOfBirth: time.Now(),
		},
		{
			Name:        "Charlie",
			Age:         3,
			ValueWeight: 5,
			Owners:      []string{"Mr.Brown", "Mrs.Brown"},
			DateOfBirth: time.Now().Add(time.Duration(-2) * time.Minute),
		},
		{
			Name:        "Charlotta",
			Age:         3,
			ValueWeight: 5,
			Owners:      []string{"Mr.Brown", "Mrs.Brown"},
		},
	}

	for _, dog := range dogs {
		err := Dogs().Add(dog)
		assert.NoError(t, err)
	}
}

func TestDogRepository_UpdateWithTTL(t *testing.T) {
	newDog := &Dog{
		Name:   "Archie",
		Age:    3,
		Owners: []string{"Ms.Johnson"},
	}
	ttl := 2 * time.Second

	err := Dogs().Add(newDog)
	assert.NoError(t, err)

	dog, err := Dogs().GetByID(newDog.Name)
	assert.NoError(t, err)
	assert.Equal(t, newDog, dog)

	newDog.Age = 5
	err = Dogs().UpdateWithTTL(newDog, ttl)

	dog, err = Dogs().GetByID(newDog.Name)
	assert.NoError(t, err)
	assert.Equal(t, newDog, dog)

	time.Sleep(ttl)

	dog, err = Dogs().GetByID(newDog.Name)
	assert.EqualError(t, err, gocql.ErrNotFound.Error())
	assert.Nil(t, dog)
}
