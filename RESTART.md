# ControlCRUD - Startup and Development Guide

## Quick Start

### Start All Services (Docker)
```bash
./START.sh
```

This will:
- Create `.env` from `.env.example` if it doesn't exist
- Generate an encryption key automatically
- Build and start all containers (PostgreSQL, Backend, Frontend, Nginx)
- Display service health status

**Access the application at:** http://autogrc.mcslab.io (or http://localhost)

### Stop All Services
```bash
./STOP.sh
```

### View Logs
```bash
docker compose logs -f           # All services
docker compose logs -f backend   # Backend only
docker compose logs -f frontend  # Frontend only
docker compose logs -f nginx     # Nginx only
docker compose logs -f postgres  # Database only
```

---

## Local Development (Without Docker)

For faster development iteration, you can run the backend and frontend locally while using Docker for PostgreSQL only.

### Prerequisites
- Go 1.22+
- Node.js 18+
- Docker (for PostgreSQL)

### Terminal 1: Start Database
```bash
docker compose up -d postgres
```

### Terminal 2: Start Backend
```bash
cd backend
export DB_HOST=localhost
export DB_PORT=5434
export DB_USER=controlcrud
export DB_PASSWORD=controlcrud_dev
export DB_NAME=controlcrud
export ENCRYPTION_KEY='mWNWBLMJNqKwEW+nk8fu+1vdYH13a0udgExw0g6Igvc='

/usr/local/go/bin/go run ./cmd/server/
```

### Terminal 3: Start Frontend
```bash
cd frontend
npm run dev
```

**Access the application at:** http://localhost:5173

The Vite dev server proxies `/api` requests to the backend at `http://localhost:8080`.

---

## Environment Configuration

### Required Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `POSTGRES_USER` | PostgreSQL username | `controlcrud` |
| `POSTGRES_PASSWORD` | PostgreSQL password | `controlcrud_dev` |
| `POSTGRES_DB` | Database name | `controlcrud` |
| `ENCRYPTION_KEY` | 32-byte base64 key for AES-256 | (auto-generated) |

### Generate New Encryption Key
```bash
openssl rand -base64 32
```

---

## Service Ports

| Service | Container Port | Host Port | Description |
|---------|---------------|-----------|-------------|
| Nginx | 80 | 80 | HTTP reverse proxy |
| Backend | 8080 | - | Go API server (internal) |
| Frontend | 80 | - | React app via Nginx (internal) |
| PostgreSQL | 5432 | 5434 | Database |

---

## Health Checks

### Check All Services
```bash
docker compose ps
```

### API Health Check
```bash
curl http://localhost/health
# or
curl http://autogrc.mcslab.io/health
```

### Database Check
```bash
docker exec controlcrud-postgres pg_isready -U controlcrud
```

---

## Troubleshooting

### Reset Everything
```bash
./STOP.sh
docker compose down -v  # Remove volumes (deletes database data!)
./START.sh
```

### Rebuild Containers
```bash
docker compose up -d --build --force-recreate
```

### Check Container Logs for Errors
```bash
docker compose logs --tail=50 backend
docker compose logs --tail=50 nginx
```

### Database Connection Issues
- Ensure PostgreSQL is healthy: `docker compose ps`
- Check port 5434 isn't blocked: `nc -zv localhost 5434`
- Verify credentials in `.env` match docker-compose defaults

### Backend Won't Start
- Check encryption key is valid base64: `echo $ENCRYPTION_KEY | base64 -d | wc -c` (should be 32)
- Verify database is accessible from backend container
- Check logs: `docker compose logs backend`

---

## DNS Configuration

To use `autogrc.mcslab.io` locally, add to `/etc/hosts`:
```
127.0.0.1 autogrc.mcslab.io
```

---

## ServiceNow Test Instance

For development testing:
- **URL:** https://dev187038.service-now.com/
- **Version:** Zurich
