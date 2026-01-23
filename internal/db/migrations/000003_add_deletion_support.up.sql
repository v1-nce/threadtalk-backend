ALTER TABLE posts ADD COLUMN deleted_at TIMESTAMP WITH TIME ZONE;
ALTER TABLE comments ADD COLUMN deleted_at TIMESTAMP WITH TIME ZONE;

CREATE INDEX idx_posts_deleted_at ON posts(deleted_at) WHERE deleted_at IS NULL;
CREATE INDEX idx_comments_deleted_at ON comments(deleted_at) WHERE deleted_at IS NULL;
