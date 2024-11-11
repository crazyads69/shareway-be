package repository

import (
	"gorm.io/gorm"
)

type IChatRepository interface {
}

type ChatRepository struct {
	db *gorm.DB
}

func NewChatRepository(db *gorm.DB) IChatRepository {
	return &ChatRepository{
		db: db,
	}
}
