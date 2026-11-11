package services

import (
	"testing"

	"github.com/simplejourney/helpdesk/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCanTransition_AllowedPaths(t *testing.T) {
	cases := []struct {
		name string
		from models.Status
		to   models.Status
		role models.Role
	}{
		{"agent open to in_progress", models.StatusOpen, models.StatusInProgress, models.RoleAgent},
		{"admin in_progress to resolved", models.StatusInProgress, models.StatusResolved, models.RoleAdmin},
		{"requester resolved to closed", models.StatusResolved, models.StatusClosed, models.RoleRequester},
		{"requester reopen resolved", models.StatusResolved, models.StatusOpen, models.RoleRequester},
		{"requester cancel open", models.StatusOpen, models.StatusClosed, models.RoleRequester},
		{"admin cancel open", models.StatusOpen, models.StatusClosed, models.RoleAdmin},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := CanTransition(tc.from, tc.to, tc.role)
			assert.NoError(t, err)
		})
	}
}

func TestCanTransition_RejectedPaths(t *testing.T) {
	cases := []struct {
		name string
		from models.Status
		to   models.Status
		role models.Role
	}{
		{"requester cannot start progress", models.StatusOpen, models.StatusInProgress, models.RoleRequester},
		{"agent cannot reopen", models.StatusResolved, models.StatusOpen, models.RoleAgent},
		{"closed cannot go to open", models.StatusClosed, models.StatusOpen, models.RoleAdmin},
		{"open cannot jump to resolved", models.StatusOpen, models.StatusResolved, models.RoleAdmin},
		{"requester cannot resolve", models.StatusInProgress, models.StatusResolved, models.RoleRequester},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := CanTransition(tc.from, tc.to, tc.role)
			require.Error(t, err)
		})
	}
}

func TestCanTransition_SameStatus(t *testing.T) {
	err := CanTransition(models.StatusOpen, models.StatusOpen, models.RoleRequester)
	assert.NoError(t, err)
}
