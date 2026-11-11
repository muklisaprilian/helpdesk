package services

import "github.com/simplejourney/helpdesk/internal/models"

// CanAccessTicket returns whether the actor may read or comment on a ticket.
func CanAccessTicket(role models.Role, actorID, requesterID string) bool {
	if role == models.RoleAgent || role == models.RoleAdmin {
		return true
	}
	return actorID == requesterID
}

// CanCreateInternalNote returns whether the actor may create an internal note.
func CanCreateInternalNote(role models.Role) bool {
	return role == models.RoleAgent || role == models.RoleAdmin
}
