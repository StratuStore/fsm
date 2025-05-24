package file

import (
	"context"
	"github.com/StratuStore/fsm/internal/fsm/service"
	"github.com/StratuStore/fsm/internal/libs/owncontext"
	"github.com/StratuStore/fsm/internal/libs/ownerrors"
	"github.com/mbretter/go-mongodb/types"
	"log/slog"
)

type Updater interface {
	Update(ctx context.Context, id types.ObjectId, size uint) error
}

type UpdateResponse struct {
	Host         string `json:"host"`
	ConnectionID string `json:"connectionID"`
}

type UpdateRequest struct {
	ID   types.ObjectId `json:"id" validate:"required"`
	Size uint           `json:"size" validate:"required"`
}

func (s *Service) Update(ctx owncontext.Context, data UpdateRequest) (*UpdateResponse, error) {
	l := s.l.With(slog.String("op", "Update"))

	file, err := s.getAndCheckUser(ctx, data.ID)
	if err != nil {
		return nil, err
	}

	err = s.s.Update(ctx, data.ID, data.Size)
	if err != nil {
		return nil, service.NewDBError(l, err)
	}

	host, connectionID, err := s.c.Update(ctx, string(file.ID))
	if err != nil {
		return nil, ownerrors.NewInternalError(l, "unable to communicate with FS", err)
	}

	return &UpdateResponse{
		Host:         host,
		ConnectionID: connectionID,
	}, nil
}
