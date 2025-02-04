package platforms

import (
	"context"

	"github.com/dapplux/twitter-haiku-bot/entities"
)

// PlatformProvider defines methods for fetching posts and posting comments.
type PlatformProvider interface {
	// FetchPosts fetches posts from the platform.
	FetchPosts(ctx context.Context, limit int) ([]entities.Post, error)
	// CommentOn posts a comment on a tweet or equivalent post.
	CommentOn(postID, message string) error
}
