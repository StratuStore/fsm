package handler

import (
	"github.com/StratuStore/fsm/internal/libs/owncontext"
	"github.com/StratuStore/fsm/internal/libs/ownerrors"
	"github.com/StratuStore/fsm/internal/libs/utils"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"log/slog"
	"net/http"
)

type Handler[T, V any] struct {
	l                    *slog.Logger
	v                    *validator.Validate
	name                 string
	input                InputFunc
	serviceWithoutResult func(owncontext.Context, *T) error
	serviceWithResult    func(owncontext.Context, *T) (*V, error)
}

func NewWithResult[T, V any](
	l *slog.Logger,
	v *validator.Validate,
	name string,
	input InputFunc,
	f func(owncontext.Context, *T) (*V, error),
) *Handler[T, V] {
	return &Handler[T, V]{
		l:                 l,
		v:                 v,
		name:              name,
		input:             input,
		serviceWithResult: f,
	}
}

func NewWithoutResult[T any](
	l *slog.Logger,
	v *validator.Validate,
	name string,
	input InputFunc,
	f func(owncontext.Context, *T) error,
) *Handler[T, any] {
	return &Handler[T, any]{
		l:                    l,
		v:                    v,
		name:                 name,
		input:                input,
		serviceWithoutResult: f,
	}
}

func (h *Handler[T, V]) Handler() func(c *fiber.Ctx) error {
	if h.serviceWithoutResult != nil {
		return h.handleWithoutResult
	}

	return h.handleWithResult
}

func (h *Handler[T, V]) handleWithResult(c *fiber.Ctx) error {
	l := h.l.With(slog.String("op", h.name))

	data, err := h.processData(l, c)
	if err != nil {
		return err
	}

	userID, err := GetUserID(l, c)
	if err != nil {
		return c.Status(http.StatusUnauthorized).JSON(utils.NewErrorResponse("authentification error"))
	}

	entity, err := h.serviceWithResult(owncontext.New(c.Context(), userID), data)
	if err != nil {
		return utils.ProcessError(l, c, err)
	}

	return c.Status(http.StatusOK).JSON(utils.NewOKResponse(entity))
}

func (h *Handler[T, V]) handleWithoutResult(c *fiber.Ctx) error {
	l := h.l.With(slog.String("op", h.name))

	data, err := h.processData(l, c)
	if err != nil {
		return err
	}

	userID, err := GetUserID(l, c)
	if err != nil {
		return c.Status(http.StatusUnauthorized).JSON(utils.NewErrorResponse("authentification error"))
	}

	err = h.serviceWithoutResult(owncontext.New(c.Context(), userID), data)
	if err != nil {
		return utils.ProcessError(l, c, err)
	}

	return c.Status(http.StatusOK).JSON(utils.NewOKResponse[any](nil))
}

func (h *Handler[T, V]) processData(l *slog.Logger, c *fiber.Ctx) (*T, error) {
	var data T
	if err := h.input(c, &data); err != nil {
		l.Error("unable to parse input data", slog.String("err", err.Error()))

		return nil, c.Status(http.StatusBadRequest).JSON(utils.NewErrorResponse("unable to parse request body"))
	}

	if err := h.v.StructCtx(c.Context(), data); err != nil {
		err := NewValidationError(err)
		l.Debug("validation error", slog.String("err", err.Error()))

		return nil, c.Status(err.Status()).JSON(utils.NewErrorResponse(err.UserMessage()))
	}

	return &data, nil
}

func GetUserID(l *slog.Logger, c *fiber.Ctx) (string, error) {
	user, ok := c.Locals("user").(*jwt.Token)
	if !ok {
		return "", ownerrors.NewUnauthorizedError(l, "unable to get user from context", "authentification error")
	}
	claims := user.Claims.(jwt.MapClaims)
	id, ok := claims["id"]
	if !ok {
		return "", ownerrors.NewUnauthorizedError(l, "unable to get id from claims", "authentification error")
	}
	idStr, ok := id.(string)
	if !ok {
		return "", ownerrors.NewUnauthorizedError(l, "unable to convert id to string", "authentification error")
	}
	_, ok = claims["jti"]
	if ok {
		return "", ownerrors.NewUnauthorizedError(l, "got refreshToken instead of access", "authentification error")
	}

	return idStr, nil
}
