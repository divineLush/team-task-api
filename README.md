# Team Task API

A REST API for managing tasks between teams. Built with Go, MySQL, and Redis.

## Quick Start

```bash
make setup   # create .env from example
make up      # build and start app + mysql + redis
```

The API is available at `http://localhost:8080`. Swagger docs at `http://localhost:8080/swagger/`.

## Configuration

All config is via environment variables (see `.env.example`):

| Variable | Default | Description |
|----------|---------|-------------|
| `SERVER_PORT` | `8080` | API listen port |
| `DB_HOST` | `localhost` | MySQL host |
| `DB_PORT` | `3306` | MySQL port |
| `DB_USER` | `teamtask` | MySQL user |
| `DB_PASSWORD` | `teamtask` | MySQL password |
| `DB_NAME` | `teamtask` | MySQL database |
| `DB_ROOT_PASSWORD` | `rootpass` | MySQL root password |
| `REDIS_ADDR` | `localhost:6379` | Redis address |
| `REDIS_PASSWORD` | | Redis password |
| `REDIS_DB` | `0` | Redis database |
| `JWT_SECRET` | `change-me-in-production` | JWT signing secret |
| `JWT_EXPIRY_HOURS` | `72` | Token expiry |

## API Endpoints

### Public

| Method | Path | Description |
|--------|------|-------------|
| `POST` | `/api/v1/register` | Create account |
| `POST` | `/api/v1/login` | Get JWT token |
| `GET` | `/health` | Health check |

### Protected (requires `Authorization: Bearer <token>`)

| Method | Path | Description |
|--------|------|-------------|
| `POST` | `/api/v1/teams` | Create team |
| `GET` | `/api/v1/teams` | List your teams |
| `GET` | `/api/v1/teams/{id}` | Get team |
| `PUT` | `/api/v1/teams/{id}` | Update team |
| `DELETE` | `/api/v1/teams/{id}` | Delete team |
| `POST` | `/api/v1/teams/{id}/invite` | Invite user to team |
| `POST` | `/api/v1/tasks` | Create task |
| `GET` | `/api/v1/tasks` | List tasks (filterable by `team_id`, `status`) |
| `GET` | `/api/v1/tasks/{id}` | Get task |
| `PUT` | `/api/v1/tasks/{id}` | Update task |
| `DELETE` | `/api/v1/tasks/{id}` | Delete task |
| `POST` | `/api/v1/tasks/{id}/comments` | Add comment |
| `GET` | `/api/v1/tasks/{id}/comments` | List comments |
| `DELETE` | `/api/v1/teams/{id}/comments/{commentID}` | Delete comment |
| `GET` | `/api/v1/teams/{id}/history` | View change history |
| `GET` | `/api/v1/teams/{id}/stats` | Team stats (owner/admin only) |

### Roles

- **owner** — full team control, can delete team and reassign tasks
- **admin** — can manage tasks and invite members
- **member** — can create/edit tasks, add comments

## Migrations

Migrations run automatically on app startup. To run manually:

```bash
make migrate-up     # apply all
make migrate-down   # rollback all
```

## Examples

### Register

```bash
curl -X POST http://localhost:8080/api/v1/register \
  -H "Content-Type: application/json" \
  -d '{"name": "Alice", "email": "alice@example.com", "password": "secret123"}'
```

### Login

```bash
curl -X POST http://localhost:8080/api/v1/login \
  -H "Content-Type: application/json" \
  -d '{"email": "alice@example.com", "password": "secret123"}'
```

### Create Team

```bash
curl -X POST http://localhost:8080/api/v1/teams \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{"name": "Backend Team"}'
```

### Create Task

```bash
curl -X POST http://localhost:8080/api/v1/tasks \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{"title": "Fix login bug", "team_id": "<team-uuid>", "assignee_id": "<user-uuid>"}'
```

### List Tasks by Status

```bash
curl "http://localhost:8080/api/v1/tasks?status=pending" \
  -H "Authorization: Bearer <token>"
```

## Development

```bash
make build          # build binary locally
make run            # run locally (needs db + redis on host)
make test           # run tests
make test-race      # run tests with race detector
make test-cover     # run with coverage
make swagger        # regenerate swagger docs
make down           # stop docker containers
make clean          # remove build artifacts
```
