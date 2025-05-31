package storage

import (
	"context"
	"fmt"
	"github.com/StratuStore/fsm/internal/fsm/core"
	"go.mongodb.org/mongo-driver/bson"
)

func (s *DirectoryStorage) GetGlobalWithPaginationAndFiltering(
	ctx context.Context,
	userID string,
	directoryFilter bson.D,
	fileFilter bson.D,
	offset, limit uint,
	sortByField string,
	sortOrder int,
) (*core.DirectoryLike, error) {
	var result core.DirectoryLike

	if directoryFilter != nil {
		cursor, err := s.db.Collection(DirectoryCollection).Aggregate(
			ctx,
			aggregationFilter(userID, directoryFilter, DirectoryCollection, offset, limit, sortByField, sortOrder),
		)
		if err != nil {
			return nil, fmt.Errorf("failed to execute aggregation: %w", err)
		}
		defer cursor.Close(ctx)

		var directories []core.DirectoryLike
		if err = cursor.All(ctx, &directories); err != nil {
			return nil, fmt.Errorf("failed to decode directory: %w", err)
		}

		if len(directories) == 0 {
			return nil, fmt.Errorf("directory not found")
		}
		directoriesResult := directories[0]
		result.DirectoriesCount = directoriesResult.DirectoriesCount
		result.Directories = directoriesResult.Directories
	}

	if fileFilter != nil {
		cursor, err := s.db.Collection(FileCollection).Aggregate(
			ctx,
			aggregationFilter(userID, fileFilter, FileCollection, offset, limit, sortByField, sortOrder),
		)
		if err != nil {
			return nil, fmt.Errorf("failed to execute aggregation: %w", err)
		}
		defer cursor.Close(ctx)

		var directories []core.DirectoryLike
		if err = cursor.All(ctx, &directories); err != nil {
			return nil, fmt.Errorf("failed to decode directory: %w", err)
		}

		if len(directories) == 0 {
			return nil, fmt.Errorf("directory not found")
		}
		directoriesResult := directories[0]
		result.FilesCount = directoriesResult.FilesCount
		result.Files = directoriesResult.Files
	}

	return &result, nil
}

func aggregationFilter(userID string, filter bson.D, name string, offset, limit uint, sortByField string, sortOrder int) []bson.D {
	filter = append(filter, bson.E{"userID", userID})
	result := []bson.D{
		{{"$match", filter}},
	}

	if name == DirectoryCollection {
		result = append(result, bson.D{{"$project", bson.M{
			"_id":               1,
			"userID":            1,
			"parentDirectoryID": 1,
			"path":              1,
			"name":              1,
			"createdAt":         1,
			"updatedAt":         1,
			"public":            1,
			"size":              1,
			"starred":           1,
		}}})
	}

	result = append(result, []bson.D{
		{{"$facet", bson.D{
			{"result", []bson.D{{{"$count", name + "Count"}}}},
			{name, []bson.D{
				{{"$skip", offset}},
				{{"$limit", limit}},
				{{"$sort", bson.D{{sortByField, sortOrder}}}},
			}},
		}}},
		{{"$project", bson.M{
			name + "Count": bson.M{"$arrayElemAt": []interface{}{"$result." + name + "Count", 0}},
			name:           1,
		}}},
	}...)

	return result
}
