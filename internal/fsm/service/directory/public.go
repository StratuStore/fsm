package directory

import (
	"context"
	"github.com/StratuStore/fsm/internal/fsm/service"
	"github.com/StratuStore/fsm/internal/libs/owncontext"
	"github.com/mbretter/go-mongodb/types"
	"log/slog"
)

type Sharer interface {
	Share(ctx context.Context, id types.ObjectId, mode bool) error
}

type PublicateRequest struct {
	ID     types.ObjectId `json:"id" validate:"required"`
	Public bool           `json:"public" validate:"-"`
}

func (s *Service) Publicate(ctx owncontext.Context, data PublicateRequest) error {
	l := s.l.With(slog.String("op", "Publicate"))

	file, err := s.s.Get(ctx, data.ID)
	if err != nil {
		return service.NewDBError(l, err)
	}
	if file.UserID != ctx.UserID() {
		return service.NewWrongUserError(l)
	}

	err = s.s.Share(ctx, data.ID, data.Public)
	if err != nil {
		return service.NewDBError(l, err)
	}

	return nil
}
