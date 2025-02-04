package repositories

import (
	"context"
	"fmt"

	"github.com/dapplux/twitter-haiku-bot/entities"
	"gorm.io/gorm"
)

// PostRepository defines methods to manipulate Post records in the database.
type PostRepository interface {
	// Create inserts a new Post record into the database.
	Create(ctx context.Context, tx *gorm.DB, post *entities.Post) error
	// SaveBatch inserts multiple Post records.
	SaveBatch(ctx context.Context, tx *gorm.DB, posts []entities.Post) error
	// FindByID retrieves a Post by its ID.
	FindByID(ctx context.Context, tx *gorm.DB, id string) (*entities.Post, error)
	// You can add other methods as needed.
}

type postRepositoryImpl struct {
	db *gorm.DB
}

func (r postRepositoryImpl) getDB(tx *gorm.DB) *gorm.DB {
	if tx == nil {
		return r.db
	}

	return tx
}

// NewPostRepository creates a new instance of PostRepository.
func NewPostRepository(db *gorm.DB) PostRepository {
	return &postRepositoryImpl{db: db}
}

// Create inserts a new Post record into the database.
// If tx is nil, the base DB is used.
func (r *postRepositoryImpl) Create(ctx context.Context, tx *gorm.DB, post *entities.Post) error {
	db := r.getDB(tx)

	return db.WithContext(ctx).Create(post).Error
}

// SaveBatch inserts multiple Post records.
func (r *postRepositoryImpl) SaveBatch(ctx context.Context, tx *gorm.DB, posts []entities.Post) error {
	db := r.getDB(tx)

	// Using Create with a slice will insert all records in one call.
	return db.WithContext(ctx).Create(&posts).Error
}

// FindByID retrieves a Post by its ID.
func (r *postRepositoryImpl) FindByID(ctx context.Context, tx *gorm.DB, id string) (*entities.Post, error) {
	var post entities.Post
	db := r.getDB(tx)

	if err := db.WithContext(ctx).First(&post, "id = ?", id).Error; err != nil {
		return nil, fmt.Errorf("failed to find post with id %s: %w", id, err)
	}
	return &post, nil
}
