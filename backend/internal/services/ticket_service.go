package services

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

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

	editingContent := req.Title != nil || req.Description != nil || req.Category != nil
	if editingContent {
		if !CanEditTicketContent(role, userID, ticket.RequesterID, ticket.Status) {
			return nil, NewAppError(http.StatusForbidden, "you cannot edit this ticket's content")
		}
		if req.Title != nil {
			title := strings.TrimSpace(*req.Title)
			if len(title) < 3 || len(title) > 200 {
				return nil, NewAppError(http.StatusBadRequest, "title must be 3-200 characters")
			}
			ticket.Title = title
		}
		if req.Description != nil {
			desc := strings.TrimSpace(*req.Description)
			if len(desc) < 3 {
				return nil, NewAppError(http.StatusBadRequest, "description must be at least 3 characters")
			}
			ticket.Description = desc
		}
		if req.Category != nil {
			cat := strings.TrimSpace(*req.Category)
			if len(cat) < 2 || len(cat) > 100 {
				return nil, NewAppError(http.StatusBadRequest, "category must be 2-100 characters")
			}
			ticket.Category = cat
		}
	}

	if req.Status != nil {
		if !req.Status.Valid() {
			return nil, NewAppError(http.StatusBadRequest, "invalid status")
		}
		if err := CanTransition(ticket.Status, *req.Status, role); err != nil {
			return nil, NewAppError(http.StatusUnprocessableEntity, err.Error())
		}
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

func (s *TicketService) Delete(id, userID string, role models.Role) error {
	ticket, err := s.tickets.GetByID(id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return NewAppError(http.StatusNotFound, "ticket not found")
		}
		return err
	}
	if !CanDeleteTicket(role, userID, ticket.RequesterID, ticket.Status) {
		return NewAppError(http.StatusForbidden, "you cannot delete this ticket")
	}
	if err := s.tickets.Delete(id); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return NewAppError(http.StatusNotFound, "ticket not found")
		}
		return err
	}
	return nil
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
