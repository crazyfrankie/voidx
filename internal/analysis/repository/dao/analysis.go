package dao

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/crazyfrankie/voidx/internal/models/entity"
)

type AnalysisDao struct {
	db *gorm.DB
}

func NewAnalysisDao(db *gorm.DB) *AnalysisDao {
	return &AnalysisDao{db: db}
}

// GetMessagesByTimeRange 根据时间范围获取消息数据
func (d *AnalysisDao) GetMessagesByTimeRange(ctx context.Context, appID uuid.UUID, startAt, endAt time.Time) ([]entity.Message, error) {
	var messages []entity.Message

	// 转换时间为毫秒时间戳
	startTimestamp := startAt.UnixMilli()
	endTimestamp := endAt.UnixMilli()

	err := d.db.WithContext(ctx).
		Preload("AgentThoughts").
		Where("app_id = ? AND ctime >= ? AND ctime < ? AND answer != ?",
			appID, startTimestamp, endTimestamp, "").
		Find(&messages).Error

	return messages, err
}

// GetApp 获取应用信息（用于权限验证）
func (d *AnalysisDao) GetApp(ctx context.Context, appID uuid.UUID) (*entity.App, error) {
	var app entity.App
	err := d.db.WithContext(ctx).Where("id = ?", appID).First(&app).Error
	if err != nil {
		return nil, err
	}
	return &app, nil
}
