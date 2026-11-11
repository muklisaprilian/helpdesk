package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRoleValid(t *testing.T) {
	assert.True(t, RoleAdmin.Valid())
	assert.True(t, RoleAgent.Valid())
	assert.True(t, RoleRequester.Valid())
	assert.False(t, Role("superuser").Valid())
}

func TestPriorityValid(t *testing.T) {
	assert.True(t, PriorityUrgent.Valid())
	assert.False(t, Priority("critical").Valid())
}

func TestStatusValid(t *testing.T) {
	assert.True(t, StatusInProgress.Valid())
	assert.False(t, Status("pending").Valid())
}
