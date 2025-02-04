package services

import (
	"context"
	"fmt"
	"log"

	"github.com/dapplux/twitter-haiku-bot/entities"
	"github.com/dapplux/twitter-haiku-bot/infrastructure/ai"
	"github.com/dapplux/twitter-haiku-bot/infrastructure/database/repositories"
	"github.com/dapplux/twitter-haiku-bot/infrastructure/platforms"
	"github.com/google/uuid"
	"github.com/guregu/null"
	"gorm.io/gorm"
)

type HaikuService struct {
	haikuRepo     repositories.HaikuRepository
	textProcessor ai.TextProcessor
	platform      platforms.PlatformProvider
	unit          repositories.UnitOfWork
}

func NewHaikuService(unit repositories.UnitOfWork, haikuRepo repositories.HaikuRepository, textProcessor ai.TextProcessor, platform platforms.PlatformProvider) *HaikuService {
	return &HaikuService{
		haikuRepo:     haikuRepo,
		textProcessor: textProcessor,
		platform:      platform,
		unit:          unit,
	}
}

func (s *HaikuService) CreateHaikuFromUnprocessedPost(ctx context.Context) error {
	post, err := s.haikuRepo.FindOldestUnprocessedPost(ctx)
	if post == nil {
		log.Printf("No unprocessed posts found: %v", err)
		return nil
	}

	if err != nil {
		return err
	}

	haiku := entities.Haiku{
		ID:     uuid.New().String(),
		State:  entities.HaikuStateCreated,
		PostID: post.ID,
		Post:   *post,
	}

	return s.haikuRepo.Create(ctx, nil, &haiku)
}

// Step 1: Process Summary Generation
func (s *HaikuService) ProcessSummary(ctx context.Context) error {
	haiku, err := s.haikuRepo.FindOldestByState(ctx, nil, entities.HaikuStateCreated)
	if haiku == nil {
		log.Printf("No created haikus found: %v", err)
		return nil
	}

	if err != nil {
		return err
	}

	haiku.State = entities.HaikuStateSummaryGetting
	if err := s.SafeUpdate(ctx, haiku, entities.HaikuStateCreated); err != nil {
		return err
	}

	summary, err := s.textProcessor.GenerateSummary(haiku.Post.Text)
	if err != nil {
		return s.markFailedAndReturn(ctx, haiku.ID, err)
	}

	haiku.Summary = null.StringFrom(summary)
	haiku.State = entities.HaikuStateSummaryGot

	return s.SafeUpdate(ctx, haiku, entities.HaikuStateSummaryGetting)
}

// Step 2: Process Haiku Generation
func (s *HaikuService) ProcessHaikuText(ctx context.Context) error {
	haiku, err := s.haikuRepo.FindOldestByState(ctx, nil, entities.HaikuStateSummaryGot)
	if haiku == nil {
		log.Printf("No haikus ready for generation: %v", err)
		return nil
	}

	haiku.State = entities.HaikuStateHaikuTextGetting
	if err := s.SafeUpdate(ctx, haiku, entities.HaikuStateSummaryGot); err != nil {
		return err
	}

	haikuText, err := s.textProcessor.GenerateHaiku(haiku.Summary.String)
	if err != nil {
		return s.markFailedAndReturn(ctx, haiku.ID, err)
	}

	haiku.Text = null.StringFrom(haikuText)
	haiku.State = entities.HaikuStateHaikuTextGot
	return s.SafeUpdate(ctx, haiku, entities.HaikuStateHaikuTextGetting)
}

// Step 3: Post Haiku to Platform
func (s *HaikuService) PostHaiku(ctx context.Context) error {
	haiku, err := s.haikuRepo.FindOldestByState(ctx, nil, entities.HaikuStateHaikuTextGot)
	if haiku == nil {
		log.Printf("No haikus ready for posting: %v", err)
		return nil
	}

	if err != nil {
		return err
	}

	haiku.State = entities.HaikuStateComenting
	if err := s.SafeUpdate(ctx, haiku, entities.HaikuStateHaikuTextGot); err != nil {
		return err
	}

	err = s.platform.CommentOn(haiku.PostID, haiku.Text.String)
	if err != nil {
		return s.markFailedAndReturn(ctx, haiku.ID, err)
	}

	haiku.State = entities.HaikuStateDone
	return s.SafeUpdate(ctx, haiku, entities.HaikuStateComenting)
}

// SafeUpdateState uses the transaction manager to safely update a Haiku's state.
func (s *HaikuService) SafeUpdate(ctx context.Context, haiku *entities.Haiku, requiredState entities.HaikuState) error {
	return s.unit.Transaction(func(tx *gorm.DB) error {
		h, err := s.haikuRepo.FindByIDForUpdate(ctx, tx, haiku.ID)
		if err != nil {
			return fmt.Errorf("failed to fetch row for update: %w", err)
		}

		if requiredState != h.State {
			return fmt.Errorf("unexpected state: required %s, got %s", requiredState, h.State)
		}

		if err := s.haikuRepo.Save(ctx, tx, haiku); err != nil {
			return fmt.Errorf("failed to save row: %w", err)
		}
		return nil
	})
}

func (s *HaikuService) markFailedAndReturn(ctx context.Context, haikuID string, originalErr error) error {
	if markErr := s.MarkAsFailed(ctx, haikuID); markErr != nil {
		return fmt.Errorf("original error: %v; also failed to mark as failed: %w", originalErr, markErr)
	}
	return originalErr
}

// Example usage: reading with no lock, then updating in transaction
func (s *HaikuService) MarkAsFailed(ctx context.Context, haikuID string) error {
	return s.unit.Transaction(func(tx *gorm.DB) error {
		// Get the row with a FOR UPDATE lock.
		h, err := s.haikuRepo.FindByIDForUpdate(ctx, tx, haikuID)
		if err != nil {
			return fmt.Errorf("failed to fetch row for update: %w", err)
		}

		h.State = entities.HaikuStateFailed
		if err := s.haikuRepo.Save(ctx, tx, h); err != nil {
			return fmt.Errorf("failed to save row: %w", err)
		}
		return nil
	})
}
