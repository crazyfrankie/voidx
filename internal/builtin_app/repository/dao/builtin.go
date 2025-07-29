package dao

import (
	"context"

	"gorm.io/gorm"

	"github.com/crazyfrankie/voidx/internal/core/builtin_apps/entities"
	"github.com/crazyfrankie/voidx/internal/models/entity"
	"github.com/crazyfrankie/voidx/pkg/consts"
)

type BuiltinDao struct {
	db *gorm.DB
}

func NewBuiltinDao(db *gorm.DB) *BuiltinDao {
	return &BuiltinDao{db: db}
}

func (d *BuiltinDao) AddBuiltinApp(ctx context.Context, builtinApp *entities.BuiltinAppEntity, app *entity.App) (*entity.App, error) {
	if err := d.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&entity.App{}).Create(app).Error; err != nil {
			return err
		}

		draftAppConfig := &entity.AppConfigVersion{
			AppID:                app.ID,
			ModelConfig:          builtinApp.ModelConfig,
			ConfigType:           consts.AppConfigTypeDraft,
			DialogRound:          builtinApp.DialogRound,
			PresetPrompt:         builtinApp.PresetPrompt,
			Tools:                builtinApp.Tools,
			RetrievalConfig:      builtinApp.RetrievalConfig,
			LongTermMemory:       builtinApp.LongTermMemory,
			OpeningStatement:     builtinApp.OpeningStatement,
			OpeningQuestions:     builtinApp.OpeningQuestions,
			SpeechToText:         builtinApp.SpeechToText,
			TextToSpeech:         builtinApp.TextToSpeech,
			ReviewConfig:         builtinApp.ReviewConfig,
			SuggestedAfterAnswer: builtinApp.SuggestedAfterAnswer,
		}
		if err := tx.Model(&entity.AppConfig{}).Create(draftAppConfig).Error; err != nil {
			return err
		}

		return nil
	}); err != nil {
		return nil, err
	}

	return app, nil
}
