package file

import (
	"context"
	"github.com/StratuStore/fsm/internal/fsm/service"
	"github.com/StratuStore/fsm/internal/libs/owncontext"
	"github.com/mbretter/go-mongodb/types"
	"log/slog"
)

type Starer interface {
	Star(ctx context.Context, id types.ObjectId) error
}

func (s *Service) Star(ctx owncontext.Context, data *GetRequest) error {
	l := s.l.With(slog.String("op", "Star"))

	file, err := s.s.Get(ctx, data.ID)
	if err != nil {
		return service.NewDBError(l, err)
	}
	if file.UserID != ctx.UserID() {
		return service.NewWrongUserError(l)
	}

	err = s.s.Star(ctx, data.ID)
	if err != nil {
		return service.NewDBError(l, err)
	}

	return nil
}
