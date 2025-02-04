package platforms

import (
	"context"
	"log"
	"time"

	"github.com/dapplux/twitter-haiku-bot/entities"
)

// TwitterMock is a mock implementation of Twitter API
type TwitterMock struct{}

// NewTwitterMock initializes a new mock provider
func NewTwitterMock() *TwitterMock {
	return &TwitterMock{}
}

// SearchPosts returns a list of mock tweets
func (tm *TwitterMock) FetchPosts(ctx context.Context, limit int) ([]entities.Post, error) {
	// Mock tweet data
	mockData := []entities.Post{
		{
			ID: "1885316710766678022",
			Author: entities.Author{
				ID:       "123",
				Username: "user123",
			},
			Text:      "Oracle unveils new AI agents to take on tech rivals, aiming to push AI innovation further. #AI #Oracle #TechCompetition #Innovation",
			Likes:     150,
			Shares:    80,
			Replies:   25,
			Platform:  entities.PlatformTwitter,
			CreatedAt: time.Now().Add(-1 * time.Hour),
		},
		{
			ID: "1885316710628286542",
			Author: entities.Author{
				ID:       "456",
				Username: "user456",
			},
			Text:      "RT @latdovietcong: Đề nghị bộ công an phải nhanh chóng vào cuộc điều tra...",
			Likes:     10,
			Shares:    5,
			Replies:   1,
			Platform:  entities.PlatformTwitter,
			CreatedAt: time.Now().Add(-2 * time.Hour),
		},
		{
			ID: "1885316710611509517",
			Author: entities.Author{
				ID:       "789",
				Username: "user789",
			},
			Text:      "解いたら1億円！どれに挑戦？\n1️⃣ リーマン予想 – 神の声 \n2️⃣ P≠NP予想 – AI革命の鍵",
			Likes:     220,
			Shares:    100,
			Replies:   40,
			Platform:  entities.PlatformTwitter,
			CreatedAt: time.Now().Add(-3 * time.Hour),
		},
	}

	// Return up to `limit` tweets
	if limit > len(mockData) {
		limit = len(mockData)
	}
	return mockData[:limit], nil
}

// CommentOnPost mocks commenting on a tweet
func (tm *TwitterMock) CommentOn(postID, message string) error {
	log.Printf("Mock Comment on Post ID %s: %s\n", postID, message)
	return nil
}
