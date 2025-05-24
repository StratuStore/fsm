package directory

import (
	"context"
	"errors"
	"github.com/StratuStore/fsm/internal/fsm/core"
	"github.com/StratuStore/fsm/internal/fsm/service"
	"github.com/StratuStore/fsm/internal/libs/owncontext"
	"github.com/mbretter/go-mongodb/types"
	"log/slog"
	"slices"
)

type Deleter interface {
	Delete(ctx context.Context, id types.ObjectId) error
	StupidDelete(ctx context.Context, id types.ObjectId) error
}

type DeleteRequest struct {
	ID types.ObjectId `json:"id" validate:"required"`
}

func (s *Service) Delete(ctx owncontext.Context, data DeleteRequest) error {
	l := s.l.With(slog.String("op", "Delete"))

	dir, err := s.s.Get(ctx, data.ID)
	if err != nil {
		return service.NewDBError(l, err)
	}

	if dir.UserID != ctx.UserID() {
		return service.NewWrongUserError(l)
	}

	err = s.s.Delete(ctx, dir.ID)
	if err != nil {
		return service.NewDBError(l, err)
	}

	go func() {
		err := s.deleteDir(context.Background(), dir.Directories, dir.Files)
		if err != nil {
			l.Error("unable to delete dir in the background", slog.String("err", err.Error()))
		}
	}()

	return nil
}

func (s *Service) deleteDir(ctx context.Context, dirs []core.Directory, files []core.File) error {
	l := s.l.With(slog.String("op", "deleteDir"))

	dirs = slices.Clone(dirs)
	files = slices.Clone(files)

	for i := 0; i < len(dirs); i++ {
		dir, err := s.s.Get(ctx, dirs[i].ID)
		if err != nil {
			return service.NewDBError(l, err)
		}

		dirs = append(dirs, dir.Directories...)
		files = append(files, dir.Files...)

	}

	var errs error
	for _, dir := range dirs {
		errors.Join(errs, s.s.StupidDelete(ctx, dir.ID))
	}
	for _, file := range files {
		errors.Join(errs, s.c.Delete(ctx, string(file.ID)))
	}
	if errs != nil {
		return service.NewDBError(l, errs)
	}

	return nil
}
