package services

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/simplejourney/helpdesk/internal/models"
	"github.com/simplejourney/helpdesk/internal/repository"
)

type AppError struct {
	Status  int
	Message string
}

func (e *AppError) Error() string {
	return e.Message
}

func NewAppError(status int, message string) *AppError {
	return &AppError{Status: status, Message: message}
}

type TicketService struct {
	tickets  *repository.TicketRepository
	comments *repository.CommentRepository
	users    *repository.UserRepository
}

func NewTicketService(
	tickets *repository.TicketRepository,
	comments *repository.CommentRepository,
	users *repository.UserRepository,
) *TicketService {
	return &TicketService{tickets: tickets, comments: comments, users: users}
}

func (s *TicketService) Create(req models.CreateTicketRequest, requesterID string) (*models.Ticket, error) {
	if !req.Priority.Valid() {
		return nil, NewAppError(http.StatusBadRequest, "priority must be one of: low, medium, high, urgent")
	}
	return s.tickets.Create(req, requesterID)
}

func (s *TicketService) List(q models.TicketListQuery, role models.Role, userID string) (*models.PaginatedTickets, error) {
	return s.tickets.List(q, role, userID)
}

func (s *TicketService) Get(id, userID string, role models.Role) (*models.Ticket, error) {
	ticket, err := s.tickets.GetByID(id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, NewAppError(http.StatusNotFound, "ticket not found")
		}
		return nil, err
	}
	if !CanAccessTicket(role, userID, ticket.RequesterID) {
		return nil, NewAppError(http.StatusForbidden, "you can only view your own tickets")
	}

	requester, err := s.users.FindByID(ticket.RequesterID)
	if err != nil {
		return nil, err
	}
	ticket.Requester = repository.BriefFromUser(requester)
	if ticket.AssigneeID != nil {
		assignee, err := s.users.FindByID(*ticket.AssigneeID)
		if err == nil {
			ticket.Assignee = repository.BriefFromUser(assignee)
		}
	}

	includeInternal := role == models.RoleAgent || role == models.RoleAdmin
	comments, err := s.comments.ListByTicket(id, includeInternal)
	if err != nil {
		return nil, err
	}
	ticket.Comments = comments
	return ticket, nil
}

func (s *TicketService) Update(id string, req models.UpdateTicketRequest, userID string, role models.Role) (*models.Ticket, error) {
	ticket, err := s.tickets.GetByID(id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, NewAppError(http.StatusNotFound, "ticket not found")
		}
		return nil, err
	}

	if !CanAccessTicket(role, userID, ticket.RequesterID) {
		return nil, NewAppError(http.StatusForbidden, "you can only update your own tickets")
	}

	if req.Status != nil {
		if !req.Status.Valid() {
			return nil, NewAppError(http.StatusBadRequest, "invalid status")
		}
		if err := CanTransition(ticket.Status, *req.Status, role); err != nil {
			return nil, NewAppError(http.StatusUnprocessableEntity, err.Error())
		}
		// Requester closing open tickets is cancel; requester reopening resolved is allowed.
		ticket.Status = *req.Status
	}

	if req.Priority != nil {
		if role == models.RoleRequester {
			return nil, NewAppError(http.StatusForbidden, "requesters cannot change priority")
		}
		if !req.Priority.Valid() {
			return nil, NewAppError(http.StatusBadRequest, "invalid priority")
		}
		ticket.Priority = *req.Priority
	}

	if req.AssigneeID != nil {
		if role == models.RoleRequester {
			return nil, NewAppError(http.StatusForbidden, "requesters cannot assign tickets")
		}
		if *req.AssigneeID == "" {
			ticket.AssigneeID = nil
		} else {
			assignee, err := s.users.FindByID(*req.AssigneeID)
			if err != nil {
				if errors.Is(err, repository.ErrNotFound) {
					return nil, NewAppError(http.StatusBadRequest, "assignee not found")
				}
				return nil, err
			}
			if assignee.Role != models.RoleAgent && assignee.Role != models.RoleAdmin {
				return nil, NewAppError(http.StatusBadRequest, "assignee must be an agent or admin")
			}
			ticket.AssigneeID = req.AssigneeID
			// Claiming an open ticket moves it to in_progress when agent/admin assigns themselves
			// only if status was not explicitly set in the same request.
			if req.Status == nil && ticket.Status == models.StatusOpen {
				ticket.Status = models.StatusInProgress
			}
		}
	}

	if err := s.tickets.Update(ticket); err != nil {
		return nil, err
	}
	return s.Get(id, userID, role)
}

func (s *TicketService) AddComment(ticketID string, req models.CreateCommentRequest, userID string, role models.Role) (*models.Comment, error) {
	ticket, err := s.tickets.GetByID(ticketID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, NewAppError(http.StatusNotFound, "ticket not found")
		}
		return nil, err
	}
	if !CanAccessTicket(role, userID, ticket.RequesterID) {
		return nil, NewAppError(http.StatusForbidden, "you can only comment on your own tickets")
	}
	if req.IsInternal && !CanCreateInternalNote(role) {
		return nil, NewAppError(http.StatusForbidden, "requesters cannot create internal notes")
	}

	comment, err := s.comments.Create(ticketID, userID, req.Body, req.IsInternal)
	if err != nil {
		return nil, err
	}
	author, err := s.users.FindByID(userID)
	if err != nil {
		return nil, err
	}
	comment.Author = repository.BriefFromUser(author)
	return comment, nil
}

type UserService struct {
	users *repository.UserRepository
}

func NewUserService(users *repository.UserRepository) *UserService {
	return &UserService{users: users}
}

func (s *UserService) List() ([]models.User, error) {
	users, err := s.users.List()
	if err != nil {
		return nil, err
	}
	for i := range users {
		users[i].PasswordHash = ""
	}
	return users, nil
}

func (s *UserService) UpdateRole(id string, role models.Role) (*models.User, error) {
	if !role.Valid() {
		return nil, NewAppError(http.StatusBadRequest, "role must be requester, agent, or admin")
	}
	user, err := s.users.UpdateRole(id, role)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, NewAppError(http.StatusNotFound, "user not found")
		}
		return nil, err
	}
	user.PasswordHash = ""
	return user, nil
}

func AsAppError(err error) *AppError {
	var ae *AppError
	if errors.As(err, &ae) {
		return ae
	}
	return NewAppError(http.StatusInternalServerError, fmt.Sprintf("internal error: %v", err))
}
