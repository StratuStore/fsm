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
	Starer
	Sharer
}

type Service struct {
	l *slog.Logger
	s Storage
	c service.Communicator
}

func New(l *slog.Logger, s Storage, c service.Communicator) *Service {
	return &Service{
		l: l.With("module", "internal.fsm.service.file.Service"),
		s: s,
		c: c,
	}
}
