package repository

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/simplejourney/helpdesk/internal/models"
)

type TicketRepository struct {
	db *sql.DB
}

func NewTicketRepository(db *sql.DB) *TicketRepository {
	return &TicketRepository{db: db}
}

func (r *TicketRepository) Create(req models.CreateTicketRequest, requesterID string) (*models.Ticket, error) {
	t := &models.Ticket{}
	var assignee sql.NullString
	err := r.db.QueryRow(`
		INSERT INTO tickets (title, description, category, priority, status, requester_id)
		VALUES ($1,$2,$3,$4,'open',$5)
		RETURNING id, title, description, category, priority, status, requester_id, assignee_id, created_at, updated_at
	`, req.Title, req.Description, req.Category, req.Priority, requesterID).Scan(
		&t.ID, &t.Title, &t.Description, &t.Category, &t.Priority, &t.Status,
		&t.RequesterID, &assignee, &t.CreatedAt, &t.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	t.AssigneeID = PtrString(assignee)
	return t, nil
}

func (r *TicketRepository) GetByID(id string) (*models.Ticket, error) {
	t := &models.Ticket{}
	var assignee sql.NullString
	err := r.db.QueryRow(`
		SELECT id, title, description, category, priority, status, requester_id, assignee_id, created_at, updated_at
		FROM tickets WHERE id=$1
	`, id).Scan(
		&t.ID, &t.Title, &t.Description, &t.Category, &t.Priority, &t.Status,
		&t.RequesterID, &assignee, &t.CreatedAt, &t.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	t.AssigneeID = PtrString(assignee)
	return t, nil
}

func (r *TicketRepository) List(q models.TicketListQuery, role models.Role, userID string) (*models.PaginatedTickets, error) {
	where := []string{"1=1"}
	args := []interface{}{}
	argN := 1

	if role == models.RoleRequester {
		where = append(where, fmt.Sprintf("t.requester_id = $%d", argN))
		args = append(args, userID)
		argN++
	}
	if q.Status != "" {
		where = append(where, fmt.Sprintf("t.status = $%d", argN))
		args = append(args, q.Status)
		argN++
	}
	if q.Priority != "" {
		where = append(where, fmt.Sprintf("t.priority = $%d", argN))
		args = append(args, q.Priority)
		argN++
	}
	if q.AssigneeID != "" {
		where = append(where, fmt.Sprintf("t.assignee_id = $%d", argN))
		args = append(args, q.AssigneeID)
		argN++
	}
	if q.Search != "" {
		where = append(where, fmt.Sprintf("t.title ILIKE $%d", argN))
		args = append(args, "%"+q.Search+"%")
		argN++
	}

	whereSQL := strings.Join(where, " AND ")
	var total int
	countSQL := "SELECT COUNT(*) FROM tickets t WHERE " + whereSQL
	if err := r.db.QueryRow(countSQL, args...).Scan(&total); err != nil {
		return nil, err
	}

	offset, perPage := PaginateOffset(q.Page, q.PerPage)
	page, totalPages := FormatPage(total, q.Page, perPage)
	col, dir := SafeSort(q.SortBy, q.SortDir)

	listArgs := append(append([]interface{}{}, args...), perPage, offset)
	listSQL := fmt.Sprintf(`
		SELECT t.id, t.title, t.description, t.category, t.priority, t.status,
		       t.requester_id, t.assignee_id, t.created_at, t.updated_at,
		       ru.id, ru.name, ru.email, ru.role,
		       au.id, au.name, au.email, au.role
		FROM tickets t
		JOIN users ru ON ru.id = t.requester_id
		LEFT JOIN users au ON au.id = t.assignee_id
		WHERE %s
		ORDER BY %s %s
		LIMIT $%d OFFSET $%d
	`, whereSQL, col, dir, argN, argN+1)

	rows, err := r.db.Query(listSQL, listArgs...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tickets := make([]models.Ticket, 0)
	for rows.Next() {
		var t models.Ticket
		var assignee sql.NullString
		var req models.UserBrief
		var aID, aName, aEmail sql.NullString
		var aRole sql.NullString
		if err := rows.Scan(
			&t.ID, &t.Title, &t.Description, &t.Category, &t.Priority, &t.Status,
			&t.RequesterID, &assignee, &t.CreatedAt, &t.UpdatedAt,
			&req.ID, &req.Name, &req.Email, &req.Role,
			&aID, &aName, &aEmail, &aRole,
		); err != nil {
			return nil, err
		}
		t.AssigneeID = PtrString(assignee)
		t.Requester = &req
		if aID.Valid {
			t.Assignee = &models.UserBrief{
				ID: aID.String, Name: aName.String, Email: aEmail.String, Role: models.Role(aRole.String),
			}
		}
		tickets = append(tickets, t)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return &models.PaginatedTickets{
		Data: tickets, Page: page, PerPage: perPage, Total: total, TotalPages: totalPages,
	}, nil
}

func (r *TicketRepository) Update(t *models.Ticket) error {
	var assignee sql.NullString
	err := r.db.QueryRow(`
		UPDATE tickets
		SET status=$2, priority=$3, assignee_id=$4, updated_at=NOW()
		WHERE id=$1
		RETURNING assignee_id, updated_at
	`, t.ID, t.Status, t.Priority, NullString(t.AssigneeID)).Scan(&assignee, &t.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrNotFound
	}
	t.AssigneeID = PtrString(assignee)
	return err
}

type CommentRepository struct {
	db *sql.DB
}

func NewCommentRepository(db *sql.DB) *CommentRepository {
	return &CommentRepository{db: db}
}

func (r *CommentRepository) Create(ticketID, authorID, body string, isInternal bool) (*models.Comment, error) {
	c := &models.Comment{}
	err := r.db.QueryRow(`
		INSERT INTO comments (ticket_id, author_id, body, is_internal)
		VALUES ($1,$2,$3,$4)
		RETURNING id, ticket_id, author_id, body, is_internal, created_at
	`, ticketID, authorID, body, isInternal).Scan(
		&c.ID, &c.TicketID, &c.AuthorID, &c.Body, &c.IsInternal, &c.CreatedAt,
	)
	return c, err
}

func (r *CommentRepository) ListByTicket(ticketID string, includeInternal bool) ([]models.Comment, error) {
	query := `
		SELECT c.id, c.ticket_id, c.author_id, c.body, c.is_internal, c.created_at,
		       u.id, u.name, u.email, u.role
		FROM comments c
		JOIN users u ON u.id = c.author_id
		WHERE c.ticket_id = $1
	`
	if !includeInternal {
		query += ` AND c.is_internal = FALSE`
	}
	query += ` ORDER BY c.created_at ASC`

	rows, err := r.db.Query(query, ticketID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	comments := make([]models.Comment, 0)
	for rows.Next() {
		var c models.Comment
		var author models.UserBrief
		if err := rows.Scan(
			&c.ID, &c.TicketID, &c.AuthorID, &c.Body, &c.IsInternal, &c.CreatedAt,
			&author.ID, &author.Name, &author.Email, &author.Role,
		); err != nil {
			return nil, err
		}
		c.Author = &author
		comments = append(comments, c)
	}
	return comments, rows.Err()
}
