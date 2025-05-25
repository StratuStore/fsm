package directory

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
	Sharer
}

type Service struct {
	l *slog.Logger
	s Storage
	c service.Communicator
}

func New(l *slog.Logger, s Storage, c service.Communicator) *Service {
	return &Service{
		l: l.With("module", "internal.fsm.service.directory.Service"),
		s: s,
		c: c,
	}
}

func isErrNotFound(err error) bool {
	return err != nil
}
