package file

import (
	"context"
	"github.com/StratuStore/fsm/internal/fsm/service"
	"github.com/StratuStore/fsm/internal/libs/owncontext"
	"github.com/StratuStore/fsm/internal/libs/ownerrors"
	"github.com/google/uuid"
	"github.com/mbretter/go-mongodb/types"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"log/slog"
)

const serviceAccountID = "fs"

type Deleter interface {
	Delete(ctx context.Context, id types.ObjectId) error
	StupidDelete(ctx context.Context, id types.ObjectId) error
}

type DeleteRequest struct {
	ID string `params:"id" validate:"required"`
}

func (s *Service) Delete(ctx owncontext.Context, data *DeleteRequest) error {
	l := s.l.With(slog.String("op", "Delete"))

	id, err := types.ObjectIdFromHex(data.ID)
	if err != nil {
		fileUUID, err := uuid.Parse(data.ID)
		if err != nil {
			return ownerrors.NewValidationError(l, "wrong id", "wrong input data", err)
		}
		var objID primitive.ObjectID
		copy(objID[:6], fileUUID[:6])
		copy(objID[6:], fileUUID[10:])

		id = types.ObjectId(objID.Hex())
	}

	file, err := s.s.Get(ctx, id)
	if err != nil {
		return service.NewDBError(l, err)
	}

	if file.UserID != ctx.UserID() && ctx.UserID() != serviceAccountID {
		return service.NewWrongUserError(l)
	}

	err = s.s.Delete(ctx, file.ID)
	if err != nil {
		return service.NewDBError(l, err)
	}

	if ctx.UserID() != serviceAccountID {
		go func() {
			err := s.c.Delete(context.Background(), file.ID)
			if err != nil {
				l.Error("unable to delete file in the background", slog.String("err", err.Error()))
			}
		}()
	}

	return nil
}
