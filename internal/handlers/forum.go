package handlers

import (
	"database/sql"
	"net/http"
	"strconv"
	"fmt"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/v1-nce/threadtalk-backend/internal/models"
)

type ForumHandler struct {
	DB *sql.DB
}

func (h *ForumHandler) CreateTopic(c *gin.Context) {
	var input models.Topic
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	query := `INSERT INTO topics (name, description) VALUES ($1, $2) RETURNING id, created_at`
	if err := h.DB.QueryRowContext(c.Request.Context(), query, input.Name, input.Description).Scan(&input.ID, &input.CreatedAt); err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Topic name already exists"})
		return
	}
	c.JSON(http.StatusCreated, input)
}

func (h *ForumHandler) CreatePost(c *gin.Context) {
	var input models.Post
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	uid, _ := c.Get("userID")
	input.UserID = uid.(int64)
	query := `INSERT INTO posts (title, content, user_id, topic_id) VALUES ($1, $2, $3, $4) RETURNING id, created_at`
	if err := h.DB.QueryRowContext(c.Request.Context(), query, input.Title, input.Content, input.UserID, input.TopicID).Scan(&input.ID, &input.CreatedAt); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create post"})
		return
	}
	c.JSON(http.StatusCreated, input)
}

func (h *ForumHandler) CreateComment(c *gin.Context) {
	var input models.Comment
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	uid, _ := c.Get("userID")
	input.UserID = uid.(int64)
	query := `INSERT INTO comments (content, user_id, post_id, parent_id) VALUES ($1, $2, $3, $4) RETURNING id, created_at`
	if err := h.DB.QueryRowContext(c.Request.Context(), query, input.Content, input.UserID, input.PostID, input.ParentID).Scan(&input.ID, &input.CreatedAt); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to post comment"})
		return
	}
	input.Children = []*models.Comment{}
	c.JSON(http.StatusCreated, input)
}

func (h *ForumHandler) GetPosts(c *gin.Context) {
	topicID := c.Param("topic_id")
	limit := 20
	cursorStr := c.Query("cursor")
	var cursor int64
	if cursorStr != "" {
		cursor, _ = strconv.ParseInt(cursorStr, 10, 64)
	}
	query := `SELECT p.id, p.title, p.content, p.created_at, u.username FROM posts p JOIN users u ON p.user_id = u.id WHERE p.topic_id = $1`
	args := []interface{}{topicID}
	if cursor > 0 {
		query += ` AND p.id < $2`
		args = append(args, cursor)
	}
	query += fmt.Sprintf(` ORDER BY p.id DESC LIMIT $%d`, len(args)+1)
	args = append(args, limit+1)
	rows, err := h.DB.QueryContext(c.Request.Context(), query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database unavailable"})
		return
	}
	defer rows.Close()
	posts := make([]models.Post, 0, limit)
	for rows.Next() {
		var p models.Post
		if err := rows.Scan(&p.ID, &p.Title, &p.Content, &p.CreatedAt, &p.Username); err == nil {
			posts = append(posts, p)
		}
	}
	if err := rows.Err(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error streaming posts"})
		return
	}
	var nextCursor string
	if len(posts) > limit {
		nextCursor = strconv.FormatInt(posts[limit-1].ID, 10)
		posts = posts[:limit]
	}
	c.JSON(http.StatusOK, gin.H{"data": posts, "next_cursor": nextCursor})
}

func (h *ForumHandler) GetPostWithComments(c *gin.Context) {
	postID := c.Param("post_id")
	var post models.Post
	var rootComments []*models.Comment
	errs := make(chan error, 2)
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		err := h.DB.QueryRowContext(c.Request.Context(), `
			SELECT p.id, p.title, p.content, p.created_at, u.username 
			FROM posts p JOIN users u ON p.user_id = u.id 
			WHERE p.id = $1`, postID).
			Scan(&post.ID, &post.Title, &post.Content, &post.CreatedAt, &post.Username)
		if err != nil {
			errs <- fmt.Errorf("post: %w", err)
		}
	}()
	go func() {
		defer wg.Done()
		rows, err := h.DB.QueryContext(c.Request.Context(), `
			SELECT c.id, c.content, c.user_id, c.parent_id, c.created_at, u.username
			FROM comments c
			JOIN users u ON c.user_id = u.id
			WHERE c.post_id = $1 
			ORDER BY c.created_at ASC`, postID)
		if err != nil {
			errs <- fmt.Errorf("comments: %w", err)
			return
		}
		defer rows.Close()
		var allComments []*models.Comment
		for rows.Next() {
			c := &models.Comment{Children: []*models.Comment{}}
			rows.Scan(&c.ID, &c.Content, &c.UserID, &c.ParentID, &c.CreatedAt, &c.Username)
			allComments = append(allComments, c)
		}
		if rows.Err() != nil {
			errs <- rows.Err()
			return
		}
		rootComments = buildCommentTree(allComments)
	}()
	wg.Wait()
	close(errs)
	for err := range errs {
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch data"})
			return
		}
	}
	c.JSON(http.StatusOK, gin.H{
		"post":     post,
		"comments": rootComments,
	})
}

func buildCommentTree(all []*models.Comment) []*models.Comment {
	lookup := make(map[int64]*models.Comment, len(all))
	var roots []*models.Comment
	for _, c := range all {
		lookup[c.ID] = c
	}
	for _, c := range all {
		if c.ParentID != nil {
			if parent, exists := lookup[*c.ParentID]; exists {
				parent.Children = append(parent.Children, c)
			} else {
			}
		} else {
			roots = append(roots, c)
		}
	}
	return roots
}

func (h *ForumHandler) GetTopics(c *gin.Context) {
    query := `SELECT id, name, description, created_at FROM topics ORDER BY name ASC`
    
    rows, err := h.DB.QueryContext(c.Request.Context(), query)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch topics"})
        return
    }
    defer rows.Close()

    topics := make([]models.Topic, 0)
    for rows.Next() {
        var t models.Topic
        if err := rows.Scan(&t.ID, &t.Name, &t.Description, &t.CreatedAt); err == nil {
            topics = append(topics, t)
        }
    }

    if err := rows.Err(); err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Error reading topics"})
        return
    }

    c.JSON(http.StatusOK, topics)
}