package directory

import (
	"context"
	"github.com/StratuStore/fsm/internal/fsm/core"
	"github.com/StratuStore/fsm/internal/fsm/service"
	"github.com/StratuStore/fsm/internal/libs/owncontext"
	"go.mongodb.org/mongo-driver/bson"
	"log/slog"
)

type Searcher interface {
	GetGlobalWithPaginationAndFiltering(
		ctx context.Context,
		userID string,
		directoryFilter bson.D,
		fileFilter bson.D,
		offset, limit uint,
		sortByField string,
		sortOrder int,
	) (*core.DirectoryLike, error)
}

type SearchRequest struct {
	Offset      uint   `query:"offset" validate:"-"`
	Limit       uint   `query:"limit" validate:"-"`
	SortByField string `query:"sortByField" validate:"-"`
	SortOrder   int    `query:"sortOrder" validate:"-"`
	core.Filter
}

func (s *Service) Search(ctx owncontext.Context, data *SearchRequest) (*core.DirectoryLike, error) {
	l := s.l.With(slog.String("op", "Get"))

	if data.Limit == 0 {
		data.Limit = DefaultLimit
	}
	if data.SortByField == "" {
		data.SortByField = DefaultSortField
	}
	if data.SortOrder == 0 {
		data.SortOrder = DefaultSortOrder
	}

	directoryFilter, fileFilter := data.Filter.ToMongoFilters()

	dir, err := s.s.GetGlobalWithPaginationAndFiltering(ctx, ctx.UserID(), directoryFilter, fileFilter, data.Offset, data.Limit, data.SortByField, data.SortOrder)
	if err != nil {
		return nil, service.NewDBError(l, err)
	}

	return dir, nil
}
