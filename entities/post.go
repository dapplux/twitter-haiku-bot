package entities

import "time"

type Post struct {
	ID        string `gorm:"primaryKey"`
	Author    Author `gorm:"type:jsonb"`
	Text      string
	Likes     int
	Shares    int
	Replies   int
	Platform  Platform
	CreatedAt time.Time
}
