package directory

import (
	"context"
	"github.com/StratuStore/fsm/internal/fsm/core"
	"github.com/StratuStore/fsm/internal/fsm/service"
	"github.com/StratuStore/fsm/internal/libs/owncontext"
	"github.com/mbretter/go-mongodb/types"
	"log/slog"
)

const (
	DefaultLimit     = 300
	DefaultSortField = "name"
	DefaultSortOrder = 1
)

type Getter interface {
	Get(ctx context.Context, id types.ObjectId) (*core.Directory, error)
	GetWithPagination(ctx context.Context, id types.ObjectId, offset, limit uint, sortByField string, sortOrder int) (*core.Directory, error)
	GetRoot(ctx context.Context, userID string, offset, limit uint, sortByField string, sortOrder int) (*core.Directory, error)
}

type GetRequest struct {
	ID          types.ObjectId `params:"id" validate:"-"`
	Offset      uint           `query:"offset" validate:"-"`
	Limit       uint           `query:"limit" validate:"-"`
	SortByField string         `query:"sortByField" validate:"-"`
	SortOrder   int            `query:"sortOrder" validate:"-"`
}

func (s *Service) Get(ctx owncontext.Context, data *GetRequest) (*core.Directory, error) {
	l := s.l.With(slog.String("op", "Get"))

	if data.Limit == 0 {
		data.Limit = DefaultLimit
	}
	if data.SortByField == "" {
		data.SortByField = DefaultSortField
	}
	if data.SortOrder == 0 {
		data.SortOrder = DefaultSortOrder
	}

	if data.ID.IsZero() {
		dir, err := s.s.GetRoot(ctx, ctx.UserID(), data.Offset, data.Limit, data.SortByField, data.SortOrder)
		if isErrNotFound(err) {
			dir, err = s.s.CreateRoot(ctx, ctx.UserID(), data.Offset, data.Limit, data.SortByField, data.SortOrder)
		}
		if err != nil {
			return nil, service.NewDBError(l, err)
		}

		return dir, nil
	}

	dir, err := s.s.GetWithPagination(ctx, data.ID, data.Offset, data.Limit, data.SortByField, data.SortOrder)
	if err != nil {
		return nil, service.NewDBError(l, err)
	}

	if dir.UserID != ctx.UserID() && !dir.Public {
		return nil, service.NewWrongUserError(l)
	}

	return dir, nil
}

func (s *Service) getAndCheckUser(ctx owncontext.Context, id types.ObjectId) (*core.Directory, error) {
	l := s.l.With(slog.String("op", "getAndCheckUser"))

	dir, err := s.s.Get(ctx, id)
	if err != nil {
		return nil, service.NewDBError(l, err)
	}

	if dir.UserID != ctx.UserID() && !dir.Public {
		return nil, service.NewWrongUserError(l)
	}

	return dir, nil
}

func (s *Service) getAndCheckOwner(ctx owncontext.Context, id types.ObjectId) (*core.Directory, error) {
	l := s.l.With(slog.String("op", "getAndCheckUser"))

	dir, err := s.s.Get(ctx, id)
	if err != nil {
		return nil, service.NewDBError(l, err)
	}

	if dir.UserID != ctx.UserID() {
		return nil, service.NewWrongUserError(l)
	}

	return dir, nil
}
