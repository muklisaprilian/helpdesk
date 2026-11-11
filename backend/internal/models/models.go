package models

import "time"

type Role string

const (
	RoleRequester Role = "requester"
	RoleAgent     Role = "agent"
	RoleAdmin     Role = "admin"
)

func (r Role) Valid() bool {
	switch r {
	case RoleRequester, RoleAgent, RoleAdmin:
		return true
	default:
		return false
	}
}

type Priority string

const (
	PriorityLow    Priority = "low"
	PriorityMedium Priority = "medium"
	PriorityHigh   Priority = "high"
	PriorityUrgent Priority = "urgent"
)

func (p Priority) Valid() bool {
	switch p {
	case PriorityLow, PriorityMedium, PriorityHigh, PriorityUrgent:
		return true
	default:
		return false
	}
}

type Status string

const (
	StatusOpen       Status = "open"
	StatusInProgress Status = "in_progress"
	StatusResolved   Status = "resolved"
	StatusClosed     Status = "closed"
)

func (s Status) Valid() bool {
	switch s {
	case StatusOpen, StatusInProgress, StatusResolved, StatusClosed:
		return true
	default:
		return false
	}
}

type User struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	Role         Role      `json:"role"`
	CreatedAt    time.Time `json:"created_at"`
}

type Ticket struct {
	ID          string     `json:"id"`
	Title       string     `json:"title"`
	Description string     `json:"description"`
	Category    string     `json:"category"`
	Priority    Priority   `json:"priority"`
	Status      Status     `json:"status"`
	RequesterID string     `json:"requester_id"`
	AssigneeID  *string    `json:"assignee_id"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	Requester   *UserBrief `json:"requester,omitempty"`
	Assignee    *UserBrief `json:"assignee,omitempty"`
	Comments    []Comment  `json:"comments,omitempty"`
}

type UserBrief struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
	Role  Role   `json:"role"`
}

type Comment struct {
	ID         string     `json:"id"`
	TicketID   string     `json:"ticket_id"`
	AuthorID   string     `json:"author_id"`
	Body       string     `json:"body"`
	IsInternal bool       `json:"is_internal"`
	CreatedAt  time.Time  `json:"created_at"`
	Author     *UserBrief `json:"author,omitempty"`
}

type RegisterRequest struct {
	Name     string `json:"name" binding:"required,min=2,max=100"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8,max=72"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type AuthResponse struct {
	Token string `json:"token"`
	User  User   `json:"user"`
}

type CreateTicketRequest struct {
	Title       string   `json:"title" binding:"required,min=3,max=200"`
	Description string   `json:"description" binding:"required,min=3"`
	Category    string   `json:"category" binding:"required,min=2,max=100"`
	Priority    Priority `json:"priority" binding:"required"`
}

type UpdateTicketRequest struct {
	Title       *string   `json:"title"`
	Description *string   `json:"description"`
	Category    *string   `json:"category"`
	Status      *Status   `json:"status"`
	Priority    *Priority `json:"priority"`
	AssigneeID  *string   `json:"assignee_id"`
}

type CreateCommentRequest struct {
	Body       string `json:"body" binding:"required,min=1"`
	IsInternal bool   `json:"is_internal"`
}

type UpdateRoleRequest struct {
	Role Role `json:"role" binding:"required"`
}

type TicketListQuery struct {
	Status     string
	Priority   string
	AssigneeID string
	Search     string
	SortBy     string
	SortDir    string
	Page       int
	PerPage    int
}

type PaginatedTickets struct {
	Data       []Ticket `json:"data"`
	Page       int      `json:"page"`
	PerPage    int      `json:"per_page"`
	Total      int      `json:"total"`
	TotalPages int      `json:"total_pages"`
}

type ErrorResponse struct {
	Error   string `json:"error"`
	Details string `json:"details,omitempty"`
}
