package entities

import (
	"time"

	"github.com/guregu/null"
)

type Haiku struct {
	ID        string `gorm:"primaryKey"`
	State     HaikuState
	Summary   null.String
	Text      null.String
	PostID    string
	Post      Post `gorm:"foreignKey:PostID"`
	CreatedAt time.Time
	UpdatedAt time.Time
}
