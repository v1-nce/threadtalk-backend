# ThreadTalk Backend

Threadtalk-backend is a production-ready REST API for a lightweight threaded discussion forum inspired by Reddit! Built with Go and optimized for AWS Lambda serverless deployment with PostgreSQL.

**Live Deployment:** [https://threadtalk-app.vercel.app](https://threadtalk-app.vercel.app)  
**Frontend Repository:** [https://github.com/v1-nce/threadtalk-frontend](https://github.com/v1-nce/threadtalk-frontend)

---

- 🔐 **Authentication** — JWT-based auth with HTTP-only cookies (Signup/Login/Logout)
- 🗂️ **Topic Management** — Create and browse discussion topics
- 📝 **Post System** — Create, view, and soft-delete posts with search
- 💬 **Threaded Comments** — Nested comment trees with unlimited depth
- ⚡ **Pagination** — Efficient cursor-based pagination for posts
- 🛡️ **Security** — BCrypt hashing, rate limiting, SQL injection prevention
- 🚀 **Lambda Ready** — Optimized connection pooling, context-based timeouts

---

## Tech Stack

| Category | Technology |
|----------|------------|
| **Framework** | Go 1.24, Gin |
| **Database** | PostgreSQL 16, pgx/v5 |
| **Auth** | JWT (golang-jwt/jwt/v5) |
| **Infrastructure** | AWS Lambda, Docker, ECR, RDS PostgreSQL |
| **CI/CD** | GitHub Actions, OIDC |

---

## Quick Start (Local Development)

### Prerequisites

- **Docker Desktop** — [Download here](https://www.docker.com/products/docker-desktop/). Make sure it's running before proceeding.

### 1. Clone and Configure

```bash
git clone https://github.com/v1-nce/threadtalk-backend.git
cd threadtalk-backend
```

Create `.env` in project root:

```env
DB_USER=postgres
DB_PASSWORD=your_password
DB_NAME=threadtalk
DB_HOST=db
DATABASE_URL=postgres://${DB_USER}:${DB_PASSWORD}@${DB_HOST}:5432/${DB_NAME}?sslmode=disable
JWT_SECRET=your-secret-key
FRONTEND_URL=http://localhost:3000
PORT=8080
```

### 2. Build and Start

```bash
docker-compose up --build
```

This builds the Docker images and starts:
- **PostgreSQL 16** — Database with persistent volume
- **Backend API** — Go server on `http://localhost:8080`

Migrations run automatically on startup. Wait for `Connected to PostgreSQL Database` in the logs.

### 3. Stop the Backend

```bash
docker-compose down
```

> **Note:** Add `-v` to also delete the database volume: `docker-compose down -v`

---

## Development Workflow

| Scenario | Command |
|----------|---------|
| **First time setup** | `docker-compose up --build` |
| **After modifying Go code** | `docker-compose up --build` |
| **Restart without code changes** | `docker-compose up` |
| **Stop all containers** | `docker-compose down` |
| **View real-time logs** | `docker-compose logs -f backend` |
| **Reset database** | `docker-compose down -v` then `docker-compose up --build` |

> **Why `--build`?** The Docker setup compiles Go to a binary. Unlike interpreted languages, Go changes require recompiling. The `--build` flag rebuilds the image with your latest code.

---

## Project Structure

```
threadtalk-backend/
├── cmd/api/main.go           # Entry point
├── internal/
│   ├── db/                   # Database connection & migrations
│   ├── handlers/             # HTTP handlers (auth, forum)
│   ├── middleware/           # Auth & rate limiting middleware
│   ├── models/               # Data models (User, Post, Comment)
│   ├── router/               # Route definitions
│   └── utils/                # JWT utilities
├── .github/workflows/        # CI/CD pipeline
├── Dockerfile                # Multi-stage build (local + lambda)
└── docker-compose.yml        # Local development setup
```

---

## API Endpoints

| Method | Endpoint | Description | Auth |
|--------|----------|-------------|------|
| `POST` | `/auth/signup` | Create account | No |
| `POST` | `/auth/login` | Login (JWT cookie) | No |
| `POST` | `/auth/logout` | Logout | No |
| `GET` | `/api/profile` | Get profile | ✅ |
| `GET` | `/topics` | List topics | No |
| `POST` | `/api/topics` | Create topic | ✅ |
| `GET` | `/topics/:id/posts` | List posts (paginated) | No |
| `POST` | `/api/posts` | Create post | ✅ |
| `GET` | `/posts/:id` | Get post with comments | No |
| `DELETE` | `/api/posts/:id` | Delete post | ✅ |
| `POST` | `/api/comments` | Create comment | ✅ |
| `DELETE` | `/api/comments/:id` | Delete comment | ✅ |
| `GET` | `/health` | Health check | No |

See [API.md](./API.md) for complete documentation with request/response examples.

---

## Environment Variables

| Variable | Description | Required |
|----------|-------------|----------|
| `DATABASE_URL` | PostgreSQL connection string | ✅ |
| `JWT_SECRET` | JWT signing key (min 32 chars recommended) | ✅ |
| `FRONTEND_URL` | CORS origin (e.g., `http://localhost:3000`) | ✅ |
| `PORT` | Server port | No (default: 8080) |

---

## Production Deployment (CI/CD)

This repository automatically deploys to **AWS Lambda** when you push to `main`.

### How It Works

1. You push code to the `main` branch
2. GitHub Actions detects the push and starts the pipeline
3. The pipeline authenticates with AWS using OIDC (no passwords stored)
4. It builds a Docker image optimized for Lambda
5. The image is pushed to Amazon ECR (container registry)
6. Lambda is updated to use the new image
7. Your changes are live!

### What You Need (First-Time Setup)

**GitHub Secrets** (Settings → Secrets → Actions):

| Secret | What It Is |
|--------|------------|
| `ECR_REPOSITORY` | Your ECR repository name (e.g., `threadtalk-backend`) |
| `LAMBDA_FUNCTION_NAME` | Your Lambda function name |

**AWS Resources:**

| Resource | Purpose |
|----------|---------|
| ECR Repository | Stores your Docker images |
| Lambda Function | Runs your API (configured for container images) |
| IAM Role | Lets GitHub authenticate via OIDC |
| RDS PostgreSQL | Production database |

### Docker Build Stages

The Dockerfile has two modes:

| Target | Command | Use Case |
|--------|---------|----------|
| `local` | `docker-compose up --build` | Local development |
| `lambda` | `docker build --target lambda .` | Production (includes Lambda adapter) |

The Lambda build adds the [AWS Lambda Web Adapter](https://github.com/awslabs/aws-lambda-web-adapter) so the same Go HTTP server works on both your machine and Lambda.

---

## Security

- **Password Hashing** — BCrypt (cost 10)
- **JWT Tokens** — 24h expiration, HTTP-only cookies
- **SQL Injection** — Parameterized queries only
- **Rate Limiting** — 1 req/s (auth), 5 req/s (public)
- **CORS** — Restricted to configured origins

---

## Contributing

Want to contribute? Here's how:

### Step 1: Fork the Repository

Click the "Fork" button on GitHub to create your own copy.

### Step 2: Clone Your Fork

```bash
git clone https://github.com/YOUR_USERNAME/threadtalk-backend.git
cd threadtalk-backend
```

### Step 3: Create a Branch

```bash
git checkout -b feature/your-feature-name
```

Use descriptive names like `feature/add-user-avatars` or `fix/login-timeout`.

### Step 4: Make Your Changes

1. Set up the local environment (see [Quick Start](#quick-start-local-development))
2. Make your code changes
3. Test locally with `docker-compose up --build`
4. Verify the API works as expected

### Step 5: Commit and Push

```bash
git add .
git commit -m "Add: description of your changes"
git push origin feature/your-feature-name
```

### Step 6: Open a Pull Request

1. Go to the original repository on GitHub
2. Click "Pull Requests" → "New Pull Request"
3. Select your fork and branch
4. Describe your changes and submit

---

## AI Declaration

Gemini 3.0 Pro was only used for research, information gathering and to tutor myself to learn new technologies.