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

// CanEditTicketContent: requester may edit own non-closed tickets; agent/admin any non-closed.
func CanEditTicketContent(role models.Role, actorID, requesterID string, status models.Status) bool {
	if status == models.StatusClosed {
		return false
	}
	if role == models.RoleAgent || role == models.RoleAdmin {
		return true
	}
	return actorID == requesterID
}

// CanDeleteTicket: requester may delete own open tickets; agent/admin any ticket.
func CanDeleteTicket(role models.Role, actorID, requesterID string, status models.Status) bool {
	if role == models.RoleAgent || role == models.RoleAdmin {
		return true
	}
	return actorID == requesterID && status == models.StatusOpen
}
