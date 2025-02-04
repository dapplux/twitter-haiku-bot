package repositories

import (
	"gorm.io/gorm"
)

type UnitOfWork interface {
	Transaction(fn func(tx *gorm.DB) error) error
}

type unitOfWorkImpl struct {
	db *gorm.DB
}

func NewUnitOfWork(db *gorm.DB) UnitOfWork {
	return &unitOfWorkImpl{db: db}
}

func (u *unitOfWorkImpl) Transaction(fn func(tx *gorm.DB) error) error {
	return u.db.Transaction(fn)
}
