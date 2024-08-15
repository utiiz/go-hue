package user

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	user := NewUser("a7db5f8a-3f8b-11ec-90d6-0242ac120003")
	assert.Equal(t, "a7db5f8a-3f8b-11ec-90d6-0242ac120003", user.String())
}

func TestString(t *testing.T) {
	user := NewUser("a7db5f8a-3f8b-11ec-90d6-0242ac120003")
	assert.Equal(t, "a7db5f8a-3f8b-11ec-90d6-0242ac120003", user.String())
}
