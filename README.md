# ThreadTalk Backend API

**threadtalk-backend** is a scalable REST API providing user authentication, threaded discussions, and pagination. It is engineered to run as a serverless container on AWS Lambda, utilizing the AWS Lambda Web Adapter for seamless portability between local development and cloud execution.

**Frontend Repository Here:** [https://github.com/v1-nce/threadtalk-frontend](https://github.com/v1-nce/threadtalk-frontend)
**Deployed Application Here:** [https://threadtalk-app.vercel.app](https://threadtalk-app.vercel.app)

## Overview

REST API providing user authentication, threaded discussions, and pagination. Optimized for AWS Lambda with connection pooling, rate limiting, and comprehensive security.

**Core Features:**
* **Authentication:** JWT-based stateless auth with HTTP-only cookies (Signup/Login/Logout)
* **Forum Management:** Create and browse topics, posts, and nested comments
* **Search:** Full-text search across post titles and content
* **Pagination:** Efficient cursor-based pagination for posts
* **Nested Comments:** Hierarchical comment trees with parent-child relationships
* **Security:** BCrypt password hashing, SQL injection prevention, input validation
* **Error Handling:** Proper HTTP status codes (400/401/404/409/500) with server-side logging
* **Rate Limiting:** IP-based rate limiting (1 req/sec for auth, 5 req/sec for public endpoints)
* **Connection Pooling:** Lambda-optimized database connection pool (5 max connections)
* **Timeout Management:** Context-based timeouts for all database operations
* **Health Checks:** `/health` endpoint for service monitoring

## Tech Stack

- **Backend:** Go 1.24, Gin, pgx/v5, JWT
- **Database:** PostgreSQL 16 with migrations
- **Infrastructure:** AWS Lambda (Docker), ECR, CloudWatch
- **CI/CD:** GitHub Actions with AWS OIDC

## Quick Start

**Prerequisites:** Docker Desktop

1. Clone and configure:
```bash
git clone https://github.com/v1-nce/threadtalk-backend.git
cd threadtalk-backend
```

2. Create `.env`:
```bash
DB_USER=postgres
DB_PASSWORD=your_password
DB_NAME=threadtalk
DB_HOST=db
DATABASE_URL=postgres://${DB_USER}:${DB_PASSWORD}@${DB_HOST}:5432/${DB_NAME}?sslmode=disable
JWT_SECRET=your-secret-key
FRONTEND_URL=http://localhost:3000
PORT=8080
```

3. Start services:
```bash
docker-compose up --build
```

API runs at `http://localhost:8080`

## Architecture
```
Client → Lambda Function URL → Lambda Web Adapter 
→ Go App (Gin) → Connection Pool → PostgreSQL
```

**Connection Pool:** 5 max, 2 idle (Lambda-optimized)

## API Endpoints

**Auth:**
- `POST /auth/signup` - Create account
- `POST /auth/login` - Login (returns JWT cookie)
- `POST /auth/logout` - Logout
- `GET /api/profile` - Get profile (protected)

**Forum:**
- `GET /topics` - List topics
- `POST /api/topics` - Create topic (protected)
- `GET /topics/:topic_id/posts` - List posts (paginated)
- `POST /api/posts` - Create post (protected)
- `GET /posts/:post_id` - Get post with comments
- `POST /api/comments` - Create comment (protected)

**System:**
- `GET /health` - Health check

See [API.md](./API.md) for complete documentation.

## Security

- BCrypt password hashing (cost 10)
- JWT tokens (24h expiration, HTTP-only cookies)
- SQL injection prevention (parameterized queries)
- Rate limiting: 1 req/s (auth), 5 req/s (public)
- Input validation on all endpoints
- CORS restricted to configured origins

## Performance

**Database:**
- Connection pool: 5 max connections
- Query timeouts: 5-10 seconds
- Indexed queries for pagination/search

**Caching Recommendations:**
- Topics: 10 minutes
- Posts: 2 minutes
- Post details: 1 minute

**Things that still need to be implemented:**
- Optimise DB hits with materialised views
- Caching with Redis
- Index optimisation

## Environment Variables

| Variable | Description | Required |
|----------|-------------|----------|
| `DATABASE_URL` | PostgreSQL connection string | Yes |
| `JWT_SECRET` | JWT signing key | Yes |
| `FRONTEND_URL` | CORS origin | Yes |
| `PORT` | Server port (default: 8080) | No |

## Project Structure
```
threadtalk-backend/
├── cmd/api/main.go              # Entry point
├── internal/
│   ├── db/                      # Database & migrations
│   ├── handlers/                # Auth & forum handlers
│   ├── middleware/              # Auth & rate limiting
│   ├── models/                  # Data models
│   ├── router/                  # Routes
│   └── utils/                   # JWT utilities
├── .github/workflows/deploy.yml # CI/CD
├── Dockerfile                   # Multi-stage build
└── docker-compose.yml           # Local development
```

## Docker

**Local:**
```bash
docker build --target local -t threadtalk-backend:local .
```

**Lambda:**
```bash
docker build --target lambda -t threadtalk-backend:lambda .
```

## Monitoring

- Structured logging to CloudWatch
- Health check at `/health`
- Database connectivity validation

## Contributing

1. Fork repository
2. Create feature branch
3. Submit pull request