package directory

import (
	"context"
	"github.com/StratuStore/fsm/internal/fsm/core"
	"github.com/StratuStore/fsm/internal/fsm/service"
	"github.com/StratuStore/fsm/internal/libs/owncontext"
	"log/slog"
)

type Creator interface {
	Create(ctx context.Context, parentDir string, name string, userID string) (*core.Directory, error)
}

type CreateRequest struct {
	ParentDir string `json:"parentDir" validate:"required"`
	Name      string `json:"name" validate:"required"`
}

func (s *Service) Create(ctx owncontext.Context, data CreateRequest) (*core.Directory, error) {
	l := s.l.With(slog.String("op", "Create"))

	dir, err := s.s.Create(ctx, data.ParentDir, data.Name, ctx.UserID())
	if err != nil {
		return nil, service.NewDBError(l, err)
	}

	return dir, nil
}

func (s *Service) initUser(ctx context.Context, userID string) (*core.Directory, error) {
	l := s.l.With(slog.String("op", "initUser"))

	dir, err := s.s.CreateRoot(ctx, userID)
	if err != nil {
		return nil, service.NewDBError(l, err)
	}

	return dir, nil
}
