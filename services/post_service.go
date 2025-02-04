package services

import (
	"context"
	"fmt"
	"log"

	"github.com/dapplux/twitter-haiku-bot/infrastructure/database/repositories"
	"github.com/dapplux/twitter-haiku-bot/infrastructure/platforms"
)

// PostService orchestrates the fetching and saving of posts.
type PostService struct {
	platform platforms.PlatformProvider
	repo     repositories.PostRepository
}

func NewPostService(repo repositories.PostRepository, platform platforms.PlatformProvider) *PostService {
	return &PostService{
		platform: platform,
		repo:     repo,
	}
}

// FetchAndSave fetches posts from all platforms and saves them in the database.
func (s *PostService) FetchAndSave(ctx context.Context, limit int) error {
	posts, err := s.platform.FetchPosts(ctx, limit)
	if err != nil {
		return fmt.Errorf("Error fetching posts from platform: %v", err)
	}

	if len(posts) == 0 {
		return fmt.Errorf("No new posts found")
	}

	if err := s.repo.SaveBatch(ctx, nil, posts); err != nil {
		return fmt.Errorf("Error saving posts: %v", err)
	}

	log.Printf("Successfully saved %d posts", len(posts))

	return nil
}
