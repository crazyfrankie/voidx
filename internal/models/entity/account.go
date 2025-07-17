package entity

import "github.com/google/uuid"

type Account struct {
	ID       uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Avatar   string    `gorm:"size:1024" json:"avatar"`
	Name     string    `gorm:"size:255;not null" json:"name"`
	Email    string    `gorm:"size:255;not null;unique" json:"email"`
	Password string    `gorm:"size:255;not null" json:"-"`
	Utime    int64     `gorm:"autoUpdateTime" json:"utime"`
	Ctime    int64     `gorm:"autoCreateTime" json:"ctime"`
}
