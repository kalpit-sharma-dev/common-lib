package template

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAcquireID(t *testing.T) {
	entity := &Entity{}
	err := entity.AcquireID()
	assert.NoError(t, err)
}
