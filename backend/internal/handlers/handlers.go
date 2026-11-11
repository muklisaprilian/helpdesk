package handlers

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/simplejourney/helpdesk/internal/middleware"
	"github.com/simplejourney/helpdesk/internal/models"
	"github.com/simplejourney/helpdesk/internal/services"
)

type Handler struct {
	auth    *services.AuthService
	tickets *services.TicketService
	users   *services.UserService
}

func New(auth *services.AuthService, tickets *services.TicketService, users *services.UserService) *Handler {
	return &Handler{auth: auth, tickets: tickets, users: users}
}

func (h *Handler) Register(c *gin.Context) {
	var req models.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "validation failed", Details: err.Error()})
		return
	}
	res, err := h.auth.Register(req)
	if err != nil {
		if err.Error() == "email already registered" {
			c.JSON(http.StatusConflict, models.ErrorResponse{Error: err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: err.Error()})
		return
	}
	c.JSON(http.StatusCreated, res)
}

func (h *Handler) Login(c *gin.Context) {
	var req models.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "validation failed", Details: err.Error()})
		return
	}
	res, err := h.auth.Login(req)
	if err != nil {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, res)
}

func (h *Handler) Me(c *gin.Context) {
	userID, _ := middleware.CurrentUser(c)
	user, err := h.auth.Me(userID)
	if err != nil {
		writeErr(c, err)
		return
	}
	c.JSON(http.StatusOK, user)
}

func (h *Handler) ListTickets(c *gin.Context) {
	userID, role := middleware.CurrentUser(c)
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "10"))
	q := models.TicketListQuery{
		Status:     c.Query("status"),
		Priority:   c.Query("priority"),
		AssigneeID: c.Query("assignee_id"),
		Search:     c.Query("search"),
		SortBy:     c.DefaultQuery("sort_by", "created_at"),
		SortDir:    c.DefaultQuery("sort_dir", "desc"),
		Page:       page,
		PerPage:    perPage,
	}
	res, err := h.tickets.List(q, role, userID)
	if err != nil {
		writeErr(c, err)
		return
	}
	c.JSON(http.StatusOK, res)
}

func (h *Handler) CreateTicket(c *gin.Context) {
	userID, _ := middleware.CurrentUser(c)
	var req models.CreateTicketRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "validation failed", Details: err.Error()})
		return
	}
	ticket, err := h.tickets.Create(req, userID)
	if err != nil {
		writeErr(c, err)
		return
	}
	c.JSON(http.StatusCreated, ticket)
}

func (h *Handler) GetTicket(c *gin.Context) {
	userID, role := middleware.CurrentUser(c)
	ticket, err := h.tickets.Get(c.Param("id"), userID, role)
	if err != nil {
		writeErr(c, err)
		return
	}
	c.JSON(http.StatusOK, ticket)
}

func (h *Handler) UpdateTicket(c *gin.Context) {
	userID, role := middleware.CurrentUser(c)
	var req models.UpdateTicketRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "validation failed", Details: err.Error()})
		return
	}
	ticket, err := h.tickets.Update(c.Param("id"), req, userID, role)
	if err != nil {
		writeErr(c, err)
		return
	}
	c.JSON(http.StatusOK, ticket)
}

func (h *Handler) DeleteTicket(c *gin.Context) {
	userID, role := middleware.CurrentUser(c)
	if err := h.tickets.Delete(c.Param("id"), userID, role); err != nil {
		writeErr(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *Handler) AddComment(c *gin.Context) {
	userID, role := middleware.CurrentUser(c)
	var req models.CreateCommentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "validation failed", Details: err.Error()})
		return
	}
	comment, err := h.tickets.AddComment(c.Param("id"), req, userID, role)
	if err != nil {
		writeErr(c, err)
		return
	}
	c.JSON(http.StatusCreated, comment)
}

func (h *Handler) ListUsers(c *gin.Context) {
	users, err := h.users.List()
	if err != nil {
		writeErr(c, err)
		return
	}
	c.JSON(http.StatusOK, users)
}

func (h *Handler) UpdateUserRole(c *gin.Context) {
	var req models.UpdateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "validation failed", Details: err.Error()})
		return
	}
	user, err := h.users.UpdateRole(c.Param("id"), req.Role)
	if err != nil {
		writeErr(c, err)
		return
	}
	c.JSON(http.StatusOK, user)
}

func (h *Handler) Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func writeErr(c *gin.Context, err error) {
	var ae *services.AppError
	if errors.As(err, &ae) {
		c.JSON(ae.Status, models.ErrorResponse{Error: ae.Message})
		return
	}
	c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: err.Error()})
}
