package app

import (
	"github.com/StratuStore/fsm/internal/fsm/communicator"
	"github.com/StratuStore/fsm/internal/fsm/handler"
	"github.com/StratuStore/fsm/internal/fsm/service"
	"github.com/StratuStore/fsm/internal/fsm/service/directory"
	"github.com/StratuStore/fsm/internal/fsm/service/file"
	"github.com/StratuStore/fsm/internal/fsm/storage"
	"github.com/StratuStore/fsm/internal/libs/config"
	"github.com/StratuStore/fsm/internal/libs/log"
	"github.com/go-playground/validator/v10"
	"go.uber.org/fx"
)

func CreateApp(cfg *config.Config) fx.Option {
	return fx.Options(
		fx.Supply(
			cfg,
		),
		fx.Provide(
			// * Common
			newValidator,
			log.New,

			// * Storage
			storage.New,
			fx.Annotate(storage.NewDirectoryStorage, fx.As(new(directory.Storage))),
			fx.Annotate(storage.NewFileStorage, fx.As(new(file.Storage))),

			// * Services
			fx.Annotate(communicator.New, fx.As(new(service.Communicator))),
			fx.Annotate(directory.New, fx.As(new(handler.DirectoryService))),
			fx.Annotate(file.New, fx.As(new(handler.FileService))),

			// * Handlers
			handler.NewDirectoryHandler,
			handler.NewFileHandler,
			handler.New,
		),
		fx.Invoke(
			startHTTPServer,
		),
	)
}

func startHTTPServer(lifecycle fx.Lifecycle, h *handler.Handler) {
	lifecycle.Append(fx.Hook{
		OnStart: h.Start,
		OnStop:  h.Stop,
	})
}

func newValidator() *validator.Validate {
	return validator.New(validator.WithRequiredStructEnabled())
}
