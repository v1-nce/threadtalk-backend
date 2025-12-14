== API Summary TLDR ==

POST /auth/signup
Request:
{
    "username": "john_doe",
    "password": "string"
}
Response:
{
    "id": 1,
    "username": "john_doe",
    "created_at": "2024-12-12T10:30:00Z",
    "updated_at": "2024-12-12T10:30:00Z"
}

POST /auth/login
Request:
{
    "username": "john_doe",
    "password": "string"
}
Response:
{
    "id": 1,
    "username": "john_doe",
    "created_at": "2024-12-12T10:30:00Z",
    "updated_at": "2024-12-12T10:30:00Z"
}

POST /auth/logout
Request:
Response:
{
    "message": "Successfully logged out"
}

GET /api/profile
Request:
Response:
{
    "id": 1,
    "username": "john_doe",
    "created_at": "2024-12-12T10:30:00Z",
    "updated_at": "2024-12-12T10:30:00Z"
}

GET /topics
Request:
Response:
[
    {
        "id": 1,
        "name": "General Discussion",
        "description": "Talk about anything",
        "created_at": "2024-12-01T10:00:00Z"
    },
    {
        "id": 2,
        "name": "Technology",
        "description": "Tech news and discussions",
        "created_at": "2024-12-01T10:05:00Z"
    }
]

POST /api/topics
Request:
{
    "name": "string",
    "description": "string"
}
Response:
{
    "id": 3,
    "name": "Gaming",
    "description": "Gaming discussions and reviews",
    "created_at": "2024-12-12T10:30:00Z"
}

GET /topics/:topic_id/posts
Request:
Query params: cursor (optional) 
Example: /topics/1/posts?cursor=81
Response:
{
    "data": [
        {
            "id": 100,
            "title": "How to learn Go?",
            "content": "I'm new to Go programming...",
            "created_at": "2024-12-12T10:30:00Z",
            "username": "john_doe"
        },
        {
            "id": 99,
            "title": "Best practices for error handling",
            "content": "What are your thoughts on...",
            "created_at": "2024-12-12T09:15:00Z",
            "username": "jane_smith"
        }
    ],
    "next_cursor": "81"
}

POST /api/posts
Request:
{
    "title": "string",
    "content": "string",
    "topic_id": 1
}
Response:
{
    "id": 101,
    "title": "How to learn Go?",
    "content": "I'm new to Go programming...",
    "user_id": 1,
    "topic_id": 1,
    "created_at": "2024-12-12T10:30:00Z"
}

GET /posts/:post_id
Request:
Example: /posts/123
Response:
{
    "post": {
        "id": 123,
        "title": "How to learn Go?",
        "content": "I'm new to programming...",
        "created_at": "2024-12-12T10:00:00Z",
        "username": "john_doe"
    },
    "comments": [
        {
            "id": 1,
            "content": "Start with the official tour!",
            "user_id": 2,
            "parent_id": null,
            "created_at": "2024-12-12T10:05:00Z",
            "username": "alice",
            "children": [
                {
                    "id": 2,
                    "content": "I second this recommendation",
                    "user_id": 3,
                    "parent_id": 1,
                    "created_at": "2024-12-12T10:10:00Z",
                    "username": "bob",
                    "children": []
                }
            ]
        },
        {
            "id": 3,
            "content": "Also check out Go by Example",
            "user_id": 4,
            "parent_id": null,
            "created_at": "2024-12-12T10:15:00Z",
            "username": "charlie",
            "children": []
        }
    ]
}

POST /api/comments
Request:
{
    "content": "string",
    "post_id": 123,
    "parent_id": null
}
Note: parent_id is null for root comments, or ID of parent comment for replies
Response:
{
    "id": 5,
    "content": "Great advice!",
    "user_id": 1,
    "post_id": 123,
    "parent_id": 1,
    "created_at": "2024-12-12T10:20:00Z",
    "children": []
}

== Dependencies Summary ==
List of Dependencies Applied:
go get github.com/jackc/pgx/v5/stdlib
go get github.com/gin-gonic/gin
go get github.com/joho/godotenv
go get github.com/golang-migrate/migrate/v4
go get github.com/golang-migrate/migrate/v4/database/postgres
go get github.com/golang-migrate/migrate/v4/source/file
go get github.com/golang-jwt/jwt/v5
go get golang.org/x/crypto/bcrypt
go get github.com/gin-contrib/cors

List of Possible Dependencies:
go get golang.org/x/time/rate
go get github.com/gorilla/sessions
go get github.com/markbates/goth
go get github.com/markbates/goth/gothic
go get github.com/markbates/goth/providers/google
