package dao

import (
	"gorm.io/gorm"
)

type AppDao struct {
	db *gorm.DB
}

func NewAppDao(db *gorm.DB) *AppDao {
	return &AppDao{db: db}
}
