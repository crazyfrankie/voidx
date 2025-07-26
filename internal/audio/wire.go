//go:build wireinject
// +build wireinject

package audio

import (
	"github.com/google/wire"
	"gorm.io/gorm"

	"github.com/crazyfrankie/voidx/internal/audio/handler"
	"github.com/crazyfrankie/voidx/internal/audio/repository"
	"github.com/crazyfrankie/voidx/internal/audio/repository/dao"
	"github.com/crazyfrankie/voidx/internal/audio/service"
)

type Handler = handler.AudioHandler

type AudioModule struct {
	Handler *Handler
}

var ProviderSet = wire.NewSet(
	dao.NewAudioDao,
	repository.NewAudioRepo,
	service.NewAudioService,
	handler.NewAudioHandler,
)

func InitAudioModule(db *gorm.DB) *AudioModule {
	wire.Build(
		ProviderSet,
		
		wire.Struct(new(AudioModule), "*"),
	)
	return new(AudioModule)
}
