package directory

import (
	"context"
	"github.com/StratuStore/fsm/internal/fsm/core"
	"github.com/StratuStore/fsm/internal/fsm/service"
	"github.com/StratuStore/fsm/internal/libs/owncontext"
	"github.com/mbretter/go-mongodb/types"
	"log/slog"
)

type Creator interface {
	Create(ctx context.Context, parentDirID types.ObjectId, userID, name string) (*core.Directory, error)
	CreateRoot(
		ctx context.Context,
		userID string,
		offset, limit uint,
		sortByField string,
		sortOrder int,
	) (*core.Directory, error)
}

type CreateRequest struct {
	ParentDirectoryID types.ObjectId `json:"parentDirectoryID" validate:"required"`
	Name              string         `json:"name" validate:"required"`
}

func (s *Service) Create(ctx owncontext.Context, data *CreateRequest) (*core.Directory, error) {
	l := s.l.With(slog.String("op", "Create"))

	dir, err := s.s.Create(ctx, data.ParentDirectoryID, ctx.UserID(), data.Name)
	if err != nil {
		return nil, service.NewDBError(l, err)
	}

	return dir, nil
}

func (s *Service) initUser(ctx context.Context, userID string, offset, limit uint, sortByField string, sortOrder int) (*core.Directory, error) {
	l := s.l.With(slog.String("op", "initUser"))

	dir, err := s.s.CreateRoot(ctx, userID, offset, limit, sortByField, sortOrder)
	if err != nil {
		return nil, service.NewDBError(l, err)
	}

	return dir, nil
}
