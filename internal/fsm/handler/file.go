package handler

import (
	"github.com/StratuStore/fsm/internal/fsm/service/file"
	"github.com/StratuStore/fsm/internal/libs/handler"
	"github.com/StratuStore/fsm/internal/libs/owncontext"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"log/slog"
)

type FileService interface {
	Get(ctx owncontext.Context, data *file.GetRequest) (*file.Response, error)
	Move(ctx owncontext.Context, data *file.MoveRequest) error
	Create(ctx owncontext.Context, data *file.CreateRequest) (*file.Response, error)
	Delete(ctx owncontext.Context, data *file.DeleteRequest) error
	Rename(ctx owncontext.Context, data *file.RenameRequest) error
	Update(ctx owncontext.Context, data *file.UpdateRequest) (*file.UpdateResponse, error)
	Star(ctx owncontext.Context, data *file.GetRequest) error
	Publicate(ctx owncontext.Context, data *file.PublicateRequest) error
}

type FileHandler struct {
	l       *slog.Logger
	v       *validator.Validate
	service FileService
}

func NewFileHandler(l *slog.Logger, v *validator.Validate, fileService FileService) *FileHandler {
	return &FileHandler{
		l:       l.With("module", "internal.fsm.handler.FileHandler"),
		v:       v,
		service: fileService,
	}
}

func (h *FileHandler) Register(app *fiber.App, subpath string) {
	api := app.Group(subpath)

	api.Get("/:id", handler.NewWithResult(h.l, h.v, "Get", handler.ParamsInput, h.service.Get).Handler())
	api.Patch("/:id/move", handler.NewWithoutResult(h.l, h.v, "Move", handler.ParamAndQueryInput, h.service.Move).Handler())
	api.Post("/", handler.NewWithResult(h.l, h.v, "Create", handler.BodyInput, h.service.Create).Handler())
	api.Delete("/:id", handler.NewWithoutResult(h.l, h.v, "Delete", handler.ParamsInput, h.service.Delete).Handler())
	api.Patch("/:id/rename", handler.NewWithoutResult(h.l, h.v, "Rename", handler.ParamAndQueryInput, h.service.Rename).Handler())
	api.Put("/:id/update", handler.NewWithResult(h.l, h.v, "Update", handler.ParamAndQueryInput, h.service.Update).Handler())
	api.Patch("/:id/star", handler.NewWithoutResult(h.l, h.v, "Star", handler.ParamsInput, h.service.Star).Handler())
	api.Patch("/:id/share", handler.NewWithoutResult(h.l, h.v, "Publicate", handler.ParamAndQueryInput, h.service.Publicate).Handler())
}
