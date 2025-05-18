package directory

import (
	"context"
	"github.com/StratuStore/fsm/internal/fsm/core"
	"github.com/StratuStore/fsm/internal/fsm/service"
	"github.com/StratuStore/fsm/internal/libs/owncontext"
	"log/slog"
)

type Getter interface {
	GetByPath(ctx context.Context, path []string) (*core.Directory, error)
	Get(ctx context.Context, id string) (*core.Directory, error)
	CreateRoot(ctx context.Context, userID string) (*core.Directory, error)
}

type GetByPathRequest struct {
	Path []string `query:"path" validate:"-"`
}

func (s *Service) GetByPath(ctx owncontext.Context, data GetByPathRequest) (*core.Directory, error) {
	l := s.l.With(slog.String("op", "GetByPath"))
	path := data.Path

	if len(path) == 0 {
		path = nil
	}

	dir, err := s.s.GetByPath(ctx, path)
	if err != nil {
		if len(path) == 0 && isErrNotFound(err) {
			return s.initUser(ctx, ctx.UserID())
		}

		return nil, service.NewDBError(l, err)
	}

	if dir.UserID != ctx.UserID() && !dir.Public {
		return nil, service.NewWrongUserError(l)
	}

	return dir, nil
}

type GetRequest struct {
	ID string `params:"id" validate:"required"`
}

func (s *Service) Get(ctx owncontext.Context, data GetRequest) (*core.Directory, error) {
	return s.getAndCheckUser(ctx, data.ID)
}

func (s *Service) getAndCheckUser(ctx owncontext.Context, id string) (*core.Directory, error) {
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
