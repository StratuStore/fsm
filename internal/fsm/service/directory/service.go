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

func isErrNotFound(err error) bool {
	return err != nil
}
