package file

import (
	"context"
	"github.com/StratuStore/fsm/internal/fsm/service"
	"github.com/StratuStore/fsm/internal/libs/owncontext"
	"github.com/mbretter/go-mongodb/types"
	"log/slog"
)

type Deleter interface {
	Delete(ctx context.Context, id types.ObjectId) error
	StupidDelete(ctx context.Context, id types.ObjectId) error
}

type DeleteRequest struct {
	ID types.ObjectId `params:"id" validate:"required"`
}

func (s *Service) Delete(ctx owncontext.Context, data *DeleteRequest) error {
	l := s.l.With(slog.String("op", "Delete"))

	file, err := s.s.Get(ctx, data.ID)
	if err != nil {
		return service.NewDBError(l, err)
	}

	if file.UserID != ctx.UserID() {
		return service.NewWrongUserError(l)
	}

	err = s.s.Delete(ctx, file.ID)
	if err != nil {
		return service.NewDBError(l, err)
	}

	go func() {
		err := s.c.Delete(context.Background(), string(file.ID))
		if err != nil {
			l.Error("unable to delete file in the background", slog.String("err", err.Error()))
		}
	}()

	return nil
}
