package file

import (
	"context"
	"github.com/StratuStore/fsm/internal/fsm/service"
	"github.com/StratuStore/fsm/internal/libs/owncontext"
	"github.com/mbretter/go-mongodb/types"
	"log/slog"
)

type Mover interface {
	Move(ctx context.Context, fileID, toID types.ObjectId) error
}

type MoveRequest struct {
	ID types.ObjectId `json:"id" validate:"required"`
	To types.ObjectId `json:"to" validate:"required"`
}

func (s *Service) Move(ctx owncontext.Context, data MoveRequest) error {
	l := s.l.With(slog.String("op", "Move"))

	to, err := s.getAndCheckUser(ctx, data.To)
	if err != nil {
		return err
	}
	file, err := s.getAndCheckUser(ctx, data.ID)
	if err != nil {
		return err
	}

	err = s.s.Move(ctx, file.ID, to.ID)
	if err != nil {
		return service.NewDBError(l, err)
	}

	return nil
}
