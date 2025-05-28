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

type Creator interface {
	Create(ctx context.Context, parentDirID types.ObjectId, userID string, name, extension string, size uint) (*core.File, error)
}

type Response struct {
	core.File
	Host         string `json:"host"`
	ConnectionID string `json:"connectionID"`
}

type CreateRequest struct {
	ParentDirID types.ObjectId `json:"parentDirID" validate:"required"`
	Name        string         `json:"name" validate:"required"`
	Extension   string         `json:"extension" validate:"-"`
	Size        uint           `json:"size" validate:"required"`
}

func (s *Service) Create(ctx owncontext.Context, data *CreateRequest) (*Response, error) {
	l := s.l.With(slog.String("op", "Create"))

	_, err := s.getAndCheckDirectory(ctx, data.ParentDirID)
	if err != nil {
		return nil, err
	}

	file, err := s.s.Create(ctx, data.ParentDirID, ctx.UserID(), data.Name, data.Extension, data.Size)
	if err != nil {
		return nil, service.NewDBError(l, err)
	}

	host, connectionID, err := s.c.Create(ctx, string(file.ID))
	if err != nil {
		return nil, ownerrors.NewInternalError(l, "unable to communicate with FS", err)
	}

	return &Response{
		File:         *file,
		Host:         host,
		ConnectionID: connectionID,
	}, nil
}
