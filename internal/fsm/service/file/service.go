package file

import (
	"github.com/StratuStore/fsm/internal/fsm/service"
	"log/slog"
)

type Storage interface {
	Getter
	Creator
	Deleter
	Renamer
	Mover
	Updater
}

type Service struct {
	l *slog.Logger
	s Storage
	c service.Communicator
}
