package model

import (
	"testing"
	"time"

	"github.com/gocql/gocql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var expectedTom = &Cat{
	ID:          tomID,
	Name:        "Tom",
	Age:         5,
	WeightValue: 8,
	Owners:      []string{"Molly"},
}

func TestCatRepository_All(t *testing.T) {
	cats, err := Cats().All()
	require.NoError(t, err)
	assert.Len(t, cats, 2)
}

func TestCatRepository_GetByID(t *testing.T) {
	tom, err := Cats().GetByID(tomID)
	require.NoError(t, err)
	assert.Equal(t, expectedTom, tom)
}

func TestCatRepository_Update(t *testing.T) {
	defer MockCats()

	tom, err := Cats().GetByID(tomID)
	require.NoError(t, err)

	tom.Name = "Tom2"
	tom.WeightValue += 2
	tom.Owners = append(tom.Owners, "Betty")

	err = Cats().Update(tom)
	require.NoError(t, err)
	tom2, err := Cats().GetByID(tomID)
	require.NoError(t, err)

	expectedCat := &Cat{
		ID:          tomID,
		Name:        "Tom2",
		Age:         5,
		WeightValue: 10,
		Owners:      []string{"Molly", "Betty"},
	}
	assert.Equal(t, expectedCat, tom2)
}

func TestCatRepository_Delete(t *testing.T) {
	defer MockCats()

	tom, err := Cats().GetByID(tomID)
	require.NoError(t, err)

	err = Cats().Delete(tom)
	require.NoError(t, err)

	tom, err = Cats().GetByID(tomID)
	assert.EqualError(t, err, gocql.ErrNotFound.Error())
	assert.Nil(t, tom)
}

func TestCatRepository_GetByIDs(t *testing.T) {
	cats, err := Cats().All()
	require.NoError(t, err)
	catIDs := make([]gocql.UUID, 2)
	for ind, cat := range cats {
		catIDs[ind] = cat.ID
	}

	cats, err = Cats().GetByIDs(catIDs...)
	require.NoError(t, err)
	assert.Len(t, cats, 2)
}

func TestCatRepository_GetRows(t *testing.T) {
	cats, err := Cats().GetRows(tomID, 5)
	require.NoError(t, err)
	assert.Len(t, cats, 1)
}

func TestCatRepository_GetByAge(t *testing.T) {
	tom, err := Cats().GetCatsByAge(5)
	require.NoError(t, err)
	assert.Equal(t, expectedTom, tom)
}

func TestCatRepository_GetAllByAge(t *testing.T) {
	cats, err := Cats().AllCatsByAge(5)
	require.NoError(t, err)
	assert.Len(t, cats, 1)
	assert.Equal(t, expectedTom, cats[0])
}

func TestCatRepository_GetByName(t *testing.T) {
	tom, err := Cats().GetCatsByName("Tom")
	require.NoError(t, err)
	assert.Equal(t, expectedTom, tom)
}

func TestCatRepository_GetAllByName(t *testing.T) {
	cats, err := Cats().AllCatsByName("Tom")
	require.NoError(t, err)
	assert.Len(t, cats, 1)
	assert.Equal(t, expectedTom, cats[0])
}

func TestCatRepository_AddWithTTL(t *testing.T) {
	newCat := &Cat{
		ID:   gocql.TimeUUID(),
		Name: "Kitiket",
		Age:  8,
	}
	ttl := 2 * time.Second

	err := Cats().AddWithTTL(newCat, ttl)
	assert.NoError(t, err)

	cats, err := Cats().GetByIDs(newCat.ID)
	assert.NoError(t, err)
	assert.Len(t, cats, 1)

	cats, err = Cats().GetRows(newCat.ID)
	assert.NoError(t, err)
	assert.Len(t, cats, 1)

	cat, err := Cats().GetCatsByAge(newCat.Age)
	assert.NoError(t, err)
	assert.Equal(t, newCat, cat)

	time.Sleep(ttl)

	cats, err = Cats().GetByIDs(newCat.ID)
	assert.NoError(t, err)
	assert.Len(t, cats, 0)

	cat, err = Cats().GetCatsByAge(newCat.Age)
	assert.EqualError(t, err, gocql.ErrNotFound.Error())
	assert.Nil(t, cat)

	cats, err = Cats().GetRows(newCat.ID)
	assert.EqualError(t, err, gocql.ErrNotFound.Error())
	assert.Len(t, cats, 0)
}
