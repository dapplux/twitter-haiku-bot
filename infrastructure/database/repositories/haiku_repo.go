package repositories

import (
	"context"
	"fmt"

	"github.com/dapplux/twitter-haiku-bot/entities"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type HaikuRepository interface {
	// Finds a haiku by ID (optionally using the transaction provided)
	FindByID(ctx context.Context, tx *gorm.DB, id string) (*entities.Haiku, error)
	// Finds a haiku with a row-level lock for update.
	FindByIDForUpdate(ctx context.Context, tx *gorm.DB, id string) (*entities.Haiku, error)
	// Saves a haiku (within the given transaction)
	Save(ctx context.Context, tx *gorm.DB, haiku *entities.Haiku) error
	FindOldestUnprocessedPost(ctx context.Context) (*entities.Post, error)
	// Create inserts a new Haiku record into the database.
	Create(ctx context.Context, tx *gorm.DB, haiku *entities.Haiku) error

	FindOldestByState(ctx context.Context, tx *gorm.DB, state entities.HaikuState) (*entities.Haiku, error)
}

type haikuRepositoryImpl struct {
	db *gorm.DB
}

func (r haikuRepositoryImpl) getDB(tx *gorm.DB) *gorm.DB {
	if tx == nil {
		return r.db
	}

	return tx
}

// NewHaikuRepository creates a new instance of HaikuRepository with an injected DB.
func NewHaikuRepository(db *gorm.DB) HaikuRepository {
	return &haikuRepositoryImpl{db: db}
}

// Create inserts a new Haiku record into the database.
// It uses the provided transaction if not nil, otherwise falls back to the base DB.
func (r *haikuRepositoryImpl) Create(ctx context.Context, tx *gorm.DB, haiku *entities.Haiku) error {
	db := r.getDB(tx)

	return db.WithContext(ctx).Create(haiku).Error
}

// FindByID uses the provided transaction (or the base DB if tx is nil) to retrieve a Haiku.
func (r *haikuRepositoryImpl) FindByID(ctx context.Context, tx *gorm.DB, id string) (*entities.Haiku, error) {
	var h entities.Haiku
	db := r.getDB(tx)

	if err := db.WithContext(ctx).First(&h, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &h, nil
}

// FindByIDForUpdate uses row-level locking (FOR UPDATE) to retrieve a Haiku.
func (r *haikuRepositoryImpl) FindByIDForUpdate(ctx context.Context, tx *gorm.DB, id string) (*entities.Haiku, error) {
	var h entities.Haiku
	db := r.getDB(tx)

	if err := db.WithContext(ctx).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		First(&h, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &h, nil
}

// Save persists the haiku within the provided transaction.
func (r *haikuRepositoryImpl) Save(ctx context.Context, tx *gorm.DB, haiku *entities.Haiku) error {
	db := r.getDB(tx)

	return db.WithContext(ctx).Save(haiku).Error
}

// FindOldestUnprocessedPost returns the oldest post that does not have an associated haiku.
// It uses a NOT EXISTS clause for efficiency.
func (r *haikuRepositoryImpl) FindOldestUnprocessedPost(ctx context.Context) (*entities.Post, error) {
	var post entities.Post
	// Using NOT EXISTS avoids the overhead of a join when checking for missing haiku records.
	err := r.db.WithContext(ctx).
		Where("NOT EXISTS (SELECT 1 FROM haikus WHERE haikus.post_id = posts.id)").
		Order("created_at ASC").
		Limit(1).
		First(&post).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("no unprocessed post found")
		}
		return nil, err
	}
	return &post, nil
}

func (r *haikuRepositoryImpl) FindOldestByState(ctx context.Context, tx *gorm.DB, state entities.HaikuState) (*entities.Haiku, error) {
	var h entities.Haiku
	db := r.getDB(tx)

	err := db.Preload("Post").WithContext(ctx).
		Where("state = ?", state).
		Order("created_at ASC").
		Limit(1).
		First(&h).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("no haiku found with state %s", state)
		}
		return nil, fmt.Errorf("failed to fetch haiku by state: %w", err)
	}
	return &h, nil
}
