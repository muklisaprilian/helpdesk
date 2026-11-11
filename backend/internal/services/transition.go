package services

import (
	"fmt"

	"github.com/simplejourney/helpdesk/internal/models"
)

// AllowedTransitions encodes the status transition matrix from the assignment brief.
// Key format: "from->to" mapped to roles that may perform the transition.
var AllowedTransitions = map[string][]models.Role{
	"open->in_progress":     {models.RoleAgent, models.RoleAdmin},
	"in_progress->resolved": {models.RoleAgent, models.RoleAdmin},
	"resolved->closed":      {models.RoleRequester, models.RoleAgent, models.RoleAdmin},
	"resolved->open":        {models.RoleRequester},
	"open->closed":          {models.RoleRequester, models.RoleAdmin},
}

func CanTransition(from, to models.Status, role models.Role) error {
	if from == to {
		return nil
	}
	key := fmt.Sprintf("%s->%s", from, to)
	allowed, ok := AllowedTransitions[key]
	if !ok {
		return fmt.Errorf("transition from %s to %s is not allowed", from, to)
	}
	for _, r := range allowed {
		if r == role {
			return nil
		}
	}
	return fmt.Errorf("role %s cannot transition ticket from %s to %s", role, from, to)
}
