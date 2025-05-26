package handler

import (
	"github.com/StratuStore/fsm/internal/fsm/core"
	"github.com/StratuStore/fsm/internal/fsm/service/directory"
	"github.com/StratuStore/fsm/internal/libs/handler"
	"github.com/StratuStore/fsm/internal/libs/owncontext"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"log/slog"
)

type DirectoryService interface {
	Create(ctx owncontext.Context, data *directory.CreateRequest) (*core.Directory, error)
	Delete(ctx owncontext.Context, data *directory.DeleteRequest) error
	Get(ctx owncontext.Context, data *directory.GetRequest) (*core.Directory, error)
	Rename(ctx owncontext.Context, data *directory.RenameRequest) error
	Move(ctx owncontext.Context, data *directory.MoveRequest) error
	Publicate(ctx owncontext.Context, data *directory.PublicateRequest) error
	Star(ctx owncontext.Context, data *directory.StarRequest) error
	Search(ctx owncontext.Context, data *directory.SearchRequest) (*core.DirectoryLike, error)
}

type DirectoryHandler struct {
	l       *slog.Logger
	v       *validator.Validate
	service DirectoryService
}

func NewDirectoryHandler(l *slog.Logger, v *validator.Validate, directoryService DirectoryService) *DirectoryHandler {
	return &DirectoryHandler{
		l:       l.With("module", "internal.fsm.handler.DirectoryHandler"),
		v:       v,
		service: directoryService,
	}
}

func (h *DirectoryHandler) Register(app *fiber.App, subpath string) {
	api := app.Group(subpath)

	api.Get("/:id", handler.NewWithResult(h.l, h.v, "Get", handler.ParamAndQueryInput, h.service.Get).Handler())
	api.Get("/search", handler.NewWithResult(h.l, h.v, "Search", handler.QueryInput, h.service.Search).Handler())
	api.Patch("/:id/move", handler.NewWithoutResult(h.l, h.v, "Move", handler.ParamAndQueryInput, h.service.Move).Handler())
	api.Post("/", handler.NewWithResult(h.l, h.v, "Create", handler.BodyInput, h.service.Create).Handler())
	api.Delete("/:id", handler.NewWithoutResult(h.l, h.v, "Delete", handler.ParamsInput, h.service.Delete).Handler())
	api.Patch("/:id/rename", handler.NewWithoutResult(h.l, h.v, "Rename", handler.ParamAndQueryInput, h.service.Rename).Handler())
	api.Patch("/:id/share", handler.NewWithoutResult(h.l, h.v, "Publicate", handler.ParamAndQueryInput, h.service.Publicate).Handler())
	api.Patch("/:id/star", handler.NewWithoutResult(h.l, h.v, "Star", handler.ParamsInput, h.service.Star).Handler())
}
