# Helpdesk — Shift Engineer Fullstack Test

Internal helpdesk monorepo: **Go (Gin) API**, **Vue 3 (Vite)** frontend, **PostgreSQL**.

## Quick start (Docker)

```bash
docker compose up --build
```

- Frontend: http://localhost:5173  
- Backend API: http://localhost:8080  
- Health: http://localhost:8080/health  

## Seed accounts

| Role | Email | Password |
|------|-------|----------|
| Admin | `admin@helpdesk.local` | `Admin123!` |
| Agent | `agent@helpdesk.local` | `Agent123!` |
| Requester | `requester@helpdesk.local` | `Requester123!` |

## Local development

### Prerequisites

- Go 1.22+
- Node.js 20+
- PostgreSQL 16 (or use only the `db` service from Compose)

### Database

```bash
docker compose up -d db
```

### Backend

```bash
cd backend
cp ../.env.example .env   # optional
export DATABASE_URL=postgres://helpdesk:helpdesk@localhost:5432/helpdesk?sslmode=disable
export JWT_SECRET=dev-secret
go run ./cmd/server
```

### Frontend

```bash
cd frontend
npm install
echo 'VITE_API_URL=http://localhost:8080/api' > .env
npm run dev
```

## Tests

```bash
cd backend
go test ./...
```

Meaningful coverage includes:

- Status transition matrix (`internal/services/transition_test.go`)
- Ticket access / internal-note restrictions (`internal/services/access_test.go`)
- Model enum validation (`internal/models/models_test.go`)

## API endpoints

| Method | Path | Access |
|--------|------|--------|
| POST | `/api/auth/register` | Public |
| POST | `/api/auth/login` | Public |
| GET | `/api/me` | Login |
| GET | `/api/tickets` | Login (requester: own tickets only) |
| POST | `/api/tickets` | Login |
| GET | `/api/tickets/{id}` | Owner / agent / admin |
| PATCH | `/api/tickets/{id}` | Role + transition rules |
| POST | `/api/tickets/{id}/comments` | Owner / agent / admin |
| GET | `/api/users` | Admin |
| PATCH | `/api/users/{id}/role` | Admin |

OpenAPI spec: [`docs/openapi.yaml`](docs/openapi.yaml)

## Status transitions

| From | To | By |
|------|----|----|
| open | in_progress | agent, admin |
| in_progress | resolved | agent, admin |
| resolved | closed | requester, agent, admin |
| resolved | open | requester |
| open | closed | requester, admin |

Invalid transitions return **422**. Cross-tenant reads by requesters return **403**.

## Project layout

```
helpdesk/
  backend/          Go API
  frontend/         Vue 3 app
  docs/openapi.yaml API documentation
  NOTES.md          Architecture decisions
  docker-compose.yml
```

## Assumptions

- Register always creates a **requester** account; agents/admins come from seed or admin role updates.
- Assigning an agent to an `open` ticket automatically moves status to `in_progress` when status is not set in the same request.
- Internal comments are returned only to agent/admin on ticket detail.
