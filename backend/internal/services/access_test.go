package services

import (
	"testing"

	"github.com/simplejourney/helpdesk/internal/models"
	"github.com/stretchr/testify/assert"
)

func TestCanAccessTicket_RequesterOwnOnly(t *testing.T) {
	assert.True(t, CanAccessTicket(models.RoleRequester, "u1", "u1"))
	assert.False(t, CanAccessTicket(models.RoleRequester, "u1", "u2"))
}

func TestCanAccessTicket_AgentAndAdminSeeAll(t *testing.T) {
	assert.True(t, CanAccessTicket(models.RoleAgent, "agent-1", "u2"))
	assert.True(t, CanAccessTicket(models.RoleAdmin, "admin-1", "u2"))
}

func TestCanCreateInternalNote(t *testing.T) {
	assert.False(t, CanCreateInternalNote(models.RoleRequester))
	assert.True(t, CanCreateInternalNote(models.RoleAgent))
	assert.True(t, CanCreateInternalNote(models.RoleAdmin))
}
