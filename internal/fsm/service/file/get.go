package file

import (
	"context"
	"github.com/StratuStore/fsm/internal/fsm/core"
	"github.com/StratuStore/fsm/internal/fsm/service"
	"github.com/StratuStore/fsm/internal/libs/owncontext"
	"github.com/StratuStore/fsm/internal/libs/ownerrors"
	"github.com/mbretter/go-mongodb/types"
	"log/slog"
)

type Getter interface {
	Get(ctx context.Context, id types.ObjectId) (*core.File, error)
	GetDirectory(ctx context.Context, id types.ObjectId) (*core.Directory, error)
}

type GetRequest struct {
	ID types.ObjectId `params:"id" validate:"required"`
}

func (s *Service) Get(ctx owncontext.Context, data *GetRequest) (*Response, error) {
	l := s.l.With(slog.String("op", "Get"))

	file, err := s.s.Get(ctx, data.ID)
	if err != nil {
		return nil, service.NewDBError(l, err)
	}

	if file.UserID != ctx.UserID() && !file.Public {
		return nil, service.NewWrongUserError(l)
	}

	host, connectionID, err := s.c.Open(ctx, file.ID)
	if err != nil {
		return nil, ownerrors.NewInternalError(l, "unable to communicate with FS", err)
	}

	return &Response{
		File:         *file,
		Host:         host,
		ConnectionID: connectionID,
	}, nil
}

func (s *Service) getAndCheckUser(ctx owncontext.Context, id types.ObjectId) (*core.File, error) {
	l := s.l.With(slog.String("op", "getAndCheckUser"))

	file, err := s.s.Get(ctx, id)
	if err != nil {
		return nil, service.NewDBError(l, err)
	}

	if file.UserID != ctx.UserID() {
		return nil, service.NewWrongUserError(l)
	}

	return file, nil
}

func (s *Service) getAndCheckDirectory(ctx owncontext.Context, id types.ObjectId) (*core.Directory, error) {
	l := s.l.With(slog.String("op", "getAndCheckDirectory"))

	dir, err := s.s.GetDirectory(ctx, id)
	if err != nil {
		return nil, service.NewDBError(l, err)
	}

	if dir.UserID != ctx.UserID() {
		return nil, service.NewWrongUserError(l)
	}

	return dir, nil
}
