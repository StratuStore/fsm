package directory

import (
	"errors"
	"github.com/StratuStore/fsm/internal/fsm/service"
	"github.com/StratuStore/fsm/internal/fsm/storage"
	"log/slog"
)

type Storage interface {
	Getter
	Creator
	Deleter
	Renamer
	Mover
}

type Service struct {
	l *slog.Logger
	s Storage
	c service.Communicator
}

func isErrNotFound(err error) bool {
	return errors.Is(err, storage.ErrNotFound)
}
