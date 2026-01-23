package handlers

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/v1-nce/threadtalk-backend/internal/models"
)

type ForumHandler struct {
	DB *sql.DB
}

func (h *ForumHandler) CreateTopic(c *gin.Context) {
	var input models.Topic
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()
	query := `INSERT INTO topics (name, description) VALUES ($1, $2) RETURNING id, created_at`
	if err := h.DB.QueryRowContext(ctx, query, input.Name, input.Description).Scan(&input.ID, &input.CreatedAt); err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			log.Printf("WARN: Request timeout creating topic: %s", input.Name)
			c.JSON(http.StatusRequestTimeout, gin.H{"error": "Request timeout"})
		} else {
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) && pgErr.Code == "23505" {
				c.JSON(http.StatusConflict, gin.H{"error": "Topic name already exists"})
				return
			}
			errStr := err.Error()
			if strings.Contains(errStr, "23505") || strings.Contains(errStr, "duplicate key value violates unique constraint") {
				c.JSON(http.StatusConflict, gin.H{"error": "Topic name already exists"})
				return
			}
			log.Printf("ERROR: Failed to create topic %s: %v", input.Name, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create topic"})
		}
		return
	}
	c.JSON(http.StatusCreated, input)
}

func (h *ForumHandler) CreatePost(c *gin.Context) {
	var input models.Post
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}
	uid, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	userID, ok := uid.(int64)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID"})
		return
	}
	input.UserID = userID
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()
	query := `INSERT INTO posts (title, content, user_id, topic_id) VALUES ($1, $2, $3, $4) RETURNING id, created_at`
	if err := h.DB.QueryRowContext(ctx, query, input.Title, input.Content, input.UserID, input.TopicID).Scan(&input.ID, &input.CreatedAt); err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			log.Printf("WARN: Request timeout creating post in topic %d by user %d", input.TopicID, input.UserID)
			c.JSON(http.StatusRequestTimeout, gin.H{"error": "Request timeout"})
		} else {
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) && pgErr.Code == "23503" {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid topic ID"})
				return
			}
			errStr := err.Error()
			if strings.Contains(errStr, "23503") || strings.Contains(errStr, "foreign key constraint") {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid topic ID"})
				return
			}
			log.Printf("ERROR: Failed to create post in topic %d by user %d: %v", input.TopicID, input.UserID, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create post"})
		}
		return
	}
	c.JSON(http.StatusCreated, input)
}

func (h *ForumHandler) CreateComment(c *gin.Context) {
	var input models.Comment
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}
	uid, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	userID, ok := uid.(int64)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID"})
		return
	}
	input.UserID = userID
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()
	query := `INSERT INTO comments (content, user_id, post_id, parent_id) VALUES ($1, $2, $3, $4) RETURNING id, created_at`
	if err := h.DB.QueryRowContext(ctx, query, input.Content, input.UserID, input.PostID, input.ParentID).Scan(&input.ID, &input.CreatedAt); err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			log.Printf("WARN: Request timeout creating comment on post %d by user %d", input.PostID, input.UserID)
			c.JSON(http.StatusRequestTimeout, gin.H{"error": "Request timeout"})
		} else {
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) && pgErr.Code == "23503" {
				if input.ParentID != nil {
					c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid post ID or parent comment ID"})
				} else {
					c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid post ID"})
				}
				return
			}
			errStr := err.Error()
			if strings.Contains(errStr, "23503") || strings.Contains(errStr, "foreign key constraint") {
				if input.ParentID != nil {
					c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid post ID or parent comment ID"})
				} else {
					c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid post ID"})
				}
				return
			}
			log.Printf("ERROR: Failed to create comment on post %d by user %d: %v", input.PostID, input.UserID, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to post comment"})
		}
		return
	}
	input.Children = []*models.Comment{}
	c.JSON(http.StatusCreated, input)
}

func (h *ForumHandler) GetTopics(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()
	query := `SELECT id, name, description, created_at FROM topics ORDER BY name ASC`
	rows, err := h.DB.QueryContext(ctx, query)
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			log.Printf("WARN: Request timeout fetching topics")
			c.JSON(http.StatusRequestTimeout, gin.H{"error": "Request timeout"})
		} else {
			log.Printf("ERROR: Failed to fetch topics: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch topics"})
		}
		return
	}
	defer rows.Close()
	topics := make([]models.Topic, 0)
	for rows.Next() {
		var t models.Topic
		if err := rows.Scan(&t.ID, &t.Name, &t.Description, &t.CreatedAt); err != nil {
			log.Printf("ERROR: Failed to scan topic row: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch topics"})
			return
		}
		topics = append(topics, t)
	}
	if err := rows.Err(); err != nil {
		log.Printf("ERROR: Error iterating topics: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch topics"})
		return
	}
	c.JSON(http.StatusOK, topics)
}

func (h *ForumHandler) GetPosts(c *gin.Context) {
	topicIDStr := c.Param("topic_id")
	topicID, err := strconv.ParseInt(topicIDStr, 10, 64)
	if err != nil || topicID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid topic ID"})
		return
	}
	limit := 20
	cursorStr := c.Query("cursor")
	search := c.Query("search")
	var cursor int64
	if cursorStr != "" {
		var parseErr error
		cursor, parseErr = strconv.ParseInt(cursorStr, 10, 64)
		if parseErr != nil || cursor <= 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid cursor parameter"})
			return
		}
	}
	query := `SELECT p.id,
		CASE WHEN p.deleted_at IS NULL THEN p.title ELSE '[deleted]' END,
		CASE WHEN p.deleted_at IS NULL THEN p.content ELSE '[deleted]' END,
		CASE WHEN p.deleted_at IS NULL THEN p.user_id ELSE 0 END,
		p.created_at,
		CASE WHEN p.deleted_at IS NULL THEN u.username ELSE '[deleted]' END,
		(SELECT COUNT(*) FROM comments WHERE post_id = p.id)
	FROM posts p
	JOIN users u ON p.user_id = u.id
	WHERE p.topic_id = $1`
	args := []interface{}{topicID}
	argPos := 2
	if search != "" {
		if len(search) > 200 {
			search = search[:200]
		}
		query += fmt.Sprintf(` AND (p.title ILIKE $%d OR p.content ILIKE $%d)`, argPos, argPos)
		args = append(args, "%"+search+"%")
		argPos++
	}
	if cursor > 0 {
		query += fmt.Sprintf(` AND p.id < $%d`, argPos)
		args = append(args, cursor)
		argPos++
	}
	query += fmt.Sprintf(` ORDER BY p.id DESC LIMIT $%d`, argPos)
	args = append(args, limit+1)
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()
	rows, err := h.DB.QueryContext(ctx, query, args...)
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			log.Printf("WARN: Request timeout fetching posts for topic %d", topicID)
			c.JSON(http.StatusRequestTimeout, gin.H{"error": "Request timeout"})
		} else {
			log.Printf("ERROR: Failed to fetch posts for topic %d: %v", topicID, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch posts"})
		}
		return
	}
	defer rows.Close()
	posts := make([]models.Post, 0, limit)
	for rows.Next() {
		var p models.Post
		if err := rows.Scan(&p.ID, &p.Title, &p.Content, &p.UserID, &p.CreatedAt, &p.Username, &p.CommentCount); err != nil {
			log.Printf("ERROR: Failed to scan post row for topic %d: %v", topicID, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch posts"})
			return
		}
		posts = append(posts, p)
	}
	if err := rows.Err(); err != nil {
		log.Printf("ERROR: Error iterating posts for topic %d: %v", topicID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch posts"})
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
	postIDStr := c.Param("post_id")
	postID, err := strconv.ParseInt(postIDStr, 10, 64)
	if err != nil || postID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid post ID"})
		return
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()
	var post models.Post
	var rootComments []*models.Comment
	errs := make(chan error, 2)
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		err := h.DB.QueryRowContext(ctx, `
			SELECT p.id,
				CASE WHEN p.deleted_at IS NULL THEN p.title ELSE '[deleted]' END,
				CASE WHEN p.deleted_at IS NULL THEN p.content ELSE '[deleted]' END,
				CASE WHEN p.deleted_at IS NULL THEN p.user_id ELSE 0 END,
				p.created_at,
				CASE WHEN p.deleted_at IS NULL THEN u.username ELSE '[deleted]' END,
				(SELECT COUNT(*) FROM comments WHERE post_id = p.id)
			FROM posts p
			JOIN users u ON p.user_id = u.id
			WHERE p.id = $1`, postID).
			Scan(&post.ID, &post.Title, &post.Content, &post.UserID, &post.CreatedAt, &post.Username, &post.CommentCount)
		if err != nil {
			if err == sql.ErrNoRows {
				errs <- fmt.Errorf("post not found: %w", err)
			} else {
				errs <- fmt.Errorf("post: %w", err)
			}
		}
	}()
	go func() {
		defer wg.Done()
		rows, err := h.DB.QueryContext(ctx, `
			SELECT c.id,
				CASE WHEN c.deleted_at IS NULL THEN c.content ELSE '[deleted]' END,
				c.user_id, c.parent_id, c.created_at,
				CASE WHEN c.deleted_at IS NULL THEN u.username ELSE '[deleted]' END
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
			if err := rows.Scan(&c.ID, &c.Content, &c.UserID, &c.ParentID, &c.CreatedAt, &c.Username); err != nil {
				errs <- fmt.Errorf("comment scan: %w", err)
				return
			}
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
			if ctx.Err() == context.DeadlineExceeded {
				log.Printf("WARN: Request timeout fetching post %d with comments", postID)
				c.JSON(http.StatusRequestTimeout, gin.H{"error": "Request timeout"})
			} else if err.Error() == "post not found: sql: no rows in result set" || err.Error() == "post: sql: no rows in result set" {
				c.JSON(http.StatusNotFound, gin.H{"error": "Post not found"})
			} else {
				log.Printf("ERROR: Failed to fetch post %d with comments: %v", postID, err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch data"})
			}
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

func (h *ForumHandler) DeletePost(c *gin.Context) {
	postID, err := strconv.ParseInt(c.Param("post_id"), 10, 64)
	if err != nil || postID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid post ID"})
		return
	}
	uid, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	userID, ok := uid.(int64)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()
	var deletedID int64
	err = h.DB.QueryRowContext(ctx,
		`UPDATE posts SET deleted_at = CURRENT_TIMESTAMP WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL RETURNING id`,
		postID, userID).Scan(&deletedID)
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			log.Printf("Request timeout deleting post %d by user %d", postID, userID)
			c.JSON(http.StatusRequestTimeout, gin.H{"error": "Request timeout"})
		} else if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Post not found"})
		} else {
			log.Printf("Failed to delete post %d by user %d: %v", postID, userID, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete post"})
		}
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *ForumHandler) DeleteComment(c *gin.Context) {
	commentID, err := strconv.ParseInt(c.Param("comment_id"), 10, 64)
	if err != nil || commentID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid comment ID"})
		return
	}
	uid, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	userID, ok := uid.(int64)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()
	var deletedID int64
	err = h.DB.QueryRowContext(ctx,
		`UPDATE comments SET deleted_at = CURRENT_TIMESTAMP WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL RETURNING id`,
		commentID, userID).Scan(&deletedID)
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			log.Printf("Request timeout deleting comment %d by user %d", commentID, userID)
			c.JSON(http.StatusRequestTimeout, gin.H{"error": "Request timeout"})
		} else if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Comment not found"})
		} else {
			log.Printf("Failed to delete comment %d by user %d: %v", commentID, userID, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete comment"})
		}
		return
	}
	c.Status(http.StatusNoContent)
}
