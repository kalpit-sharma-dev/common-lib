package main

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetHelpTables(t *testing.T) {
	res := getHelpTables("[cats_by_age:\"Age,ID\"]")
	var expected []*viewTable
	assert.IsType(t, expected, res)
	assert.Len(t, res, 1)
	assert.Equal(t, "cats_by_age", res[0].Name)
	assert.ElementsMatch(t, []string{"AgeColumn", "IDColumn"}, res[0].Keys)
}

func TestGetKeys(t *testing.T) {
	keyStr := "\"first,second\""
	keys := getKeys(keyStr)
	assert.ElementsMatch(t, []string{"first", "second"}, keys)
}

func TestGetTableNameAndKeys(t *testing.T) {
	tableInfo := "cats_by_age:\"first,second\""
	part, keys := getTableNameAndKeys(tableInfo)
	assert.Equal(t, "cats_by_age", part)
	assert.ElementsMatch(t, []string{"firstColumn", "secondColumn"}, keys)

	assert.Panics(t, func() {
		errTableInfo := "cats_by_age\"first,second\""
		part, keys = getTableNameAndKeys(errTableInfo)
	})
}

func TestPopulateSubstitutions(t *testing.T) {
	expectedAllSubs := "first=firstVal First=FirstVal second=secondVal Second=SecondVal IDType=UUID"
	allSubs, entityName := populateSubstitutions([]string{"first=firstVal", "second=secondVal"}, "UUID")
	assert.Equal(t, expectedAllSubs, allSubs)
	assert.Empty(t, entityName)
}

func TestAddViewFunctions(t *testing.T) {
	buffer := &bytes.Buffer{}
	viewTables := getHelpTables("[cats_by_age:\"Age,ID\"]")
	tableTempls := make([]*viewsTableTempl, len(viewTables))
	for i, viewTable := range viewTables {
		tableTempls[i] = getViewsTableTempl(viewTable, "")
	}
	err := addViewFunctions(buffer, tableTempls)
	assert.NoError(t, err)
}

func TestWriteToFile(t *testing.T) {
	assert.NotPanics(t, func() {
		writeToFile("test", nil)
	})
}
