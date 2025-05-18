package directory

import (
	"context"
	"github.com/StratuStore/fsm/internal/fsm/service"
	"github.com/StratuStore/fsm/internal/libs/owncontext"
	"log/slog"
)

type Renamer interface {
	Rename(ctx context.Context, id string, newName string) error
}

type RenameRequest struct {
	ID   string `json:"id" validate:"required"`
	Name string `json:"name" validate:"required"`
}

func (s *Service) Rename(ctx owncontext.Context, data RenameRequest) error {
	l := s.l.With(slog.String("op", "Rename"))

	dir, err := s.getAndCheckUser(ctx, data.ID)
	if err != nil {
		return err
	}

	err = s.s.Rename(ctx, dir.ID, data.Name)
	if err != nil {
		return service.NewDBError(l, err)
	}

	return nil
}
