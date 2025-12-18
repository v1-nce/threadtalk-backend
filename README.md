# ThreadTalk Backend API

> A high-performance, serverless forum backend built with Go and PostgreSQL.

**ThreadTalk** is a scalable REST API providing user authentication, threaded discussions, and pagination. It is engineered to run as a serverless container on AWS Lambda, utilizing the AWS Lambda Web Adapter for seamless portability between local development and cloud execution.

## 🚀 Key Features
* **Authentication:** JWT-based stateless auth (Signup/Login).
* **Forum Core:** Topics, Posts, and nested Comments.
* **Architecture:** Serverless (Scale-to-Zero) with AWS Lambda.
* **Security:** Production-grade VPC isolation and BCrypt password hashing.
* **DevOps:** Automated CI/CD pipeline via GitHub Actions.

## 🛠 Tech Stack
* **Language:** Go (Gin Framework)
* **Database:** PostgreSQL 15 (AWS RDS)
* **Infrastructure:** AWS Lambda (Docker Image), ECR, CloudWatch
* **Tooling:** Docker, GitHub Actions, `golang-migrate`

## 🏗 Cloud Architecture

```mermaid
graph TD
    User["User / Frontend"] -->|HTTPS Requests| FuncURL["Lambda Function URL"]
    
    subgraph "AWS Cloud (Region: ap-southeast-2)"
        FuncURL -->|JSON Event| WebAdapter["AWS Lambda Web Adapter"]
        
        subgraph "VPC (Virtual Private Cloud)"
            subgraph "Lambda Container"
                WebAdapter -->|HTTP localhost:8080| GoApp["Go Backend (Gin)"]
            end
            
            GoApp -->|TCP :5432| RDS[("Amazon RDS PostgreSQL")]
        end
    end
    
    style User fill:#f9f,stroke:#333,stroke-width:2px
    style RDS fill:#3f3,stroke:#333,stroke-width:2px
    style GoApp fill:#3f3,stroke:#333,stroke-width:2px
```

## ⚡ Getting Started (Local Development)

Follow these steps to run the backend and database on your own machine.

### Prerequisites
* **Docker Desktop** (Running)
* **Go 1.23+** (Optional, if running outside Docker)

### Step 1: Clone the Repository
```bash
git clone [https://github.com/v1-nce/threadtalk-backend.git](https://github.com/v1-nce/threadtalk-backend.git)
cd threadtalk-backend
```

### Step 2: Configure Environment
Create a `.env` file in the root directory:
```bash
DB_USER=
DB_PASSWORD=
DB_NAME=
DB_HOST=db
DATABASE_URL=postgres://${DB_USER}:${DB_PASSWORD}@${DB_HOST}:5432/${DB_NAME}?sslmode=disable
JWT_SECRET=
FRONTEND_URL=http://localhost:3000
BACKEND_URL=http://localhost:8080
PORT=8080
```

### Step 3: Start Services
Run the entire stack (Go Backend + Postgres Database) with one command:
```bash
docker-compose up --build
```

### Step 4: Verify
* **API:** Accessible at `http://localhost:8080`
* **Database:** Accessible on port `5432`

To stop the services, press `Ctrl+C` or run:
```bash
docker-compose down
```