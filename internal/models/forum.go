package models

import (
	"time"
)

type Topic struct {
	ID          int64     `json:"id"`
	Name        string    `json:"name" binding:"required,max=50"`
	Description string    `json:"description" binding:"max=600"`
	CreatedAt   time.Time `json:"created_at"`
}

type Post struct {
	ID        int64     `json:"id,string"`
	Title     string    `json:"title" binding:"required,min=5,max=250"`
	Content   string    `json:"content" binding:"max=600"`
	UserID    int64     `json:"user_id,string"`
	TopicID   int64     `json:"topic_id,string" binding:"required"`
	CreatedAt time.Time `json:"created_at"`
	Username  string    `json:"username,omitempty"`
	CommentCount int `json:"comment_count"`
}

type Comment struct {
	ID        int64      `json:"id,string"`
	Content   string     `json:"content" binding:"required,max=2000"`
	UserID    int64      `json:"user_id,string"`
	PostID    int64      `json:"post_id,string" binding:"required"`
	ParentID  *int64     `json:"parent_id,string"`
	CreatedAt time.Time  `json:"created_at"`
	Username  string     `json:"username,omitempty"`
	Children  []*Comment `json:"children,omitempty"`
}
