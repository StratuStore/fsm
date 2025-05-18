package directory

import (
	"context"
	"github.com/StratuStore/fsm/internal/fsm/service"
	"github.com/StratuStore/fsm/internal/libs/owncontext"
	"log/slog"
)

type Mover interface {
	Move(ctx context.Context, dirID, fromID, toID string) error
}

type MoveRequest struct {
	ID   string `json:"id" validate:"required"`
	From string `json:"from" validate:"required"`
	To   string `json:"to" validate:"required"`
}

func (s *Service) Move(ctx owncontext.Context, data MoveRequest) error {
	l := s.l.With(slog.String("op", "Move"))

	from, err := s.getAndCheckUser(ctx, data.From)
	if err != nil {
		return err
	}
	to, err := s.getAndCheckUser(ctx, data.To)
	if err != nil {
		return err
	}
	dir, err := s.getAndCheckUser(ctx, data.ID)
	if err != nil {
		return err
	}

	err = s.s.Move(ctx, dir.ID, from.ID, to.ID)
	if err != nil {
		return service.NewDBError(l, err)
	}

	return nil
}
