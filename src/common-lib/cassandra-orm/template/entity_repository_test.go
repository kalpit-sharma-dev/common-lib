package template

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEntityRepo_Entities(t *testing.T) {
	entities := Entities()
	assert.IsType(t, new(EntityRepository), entities)
}
