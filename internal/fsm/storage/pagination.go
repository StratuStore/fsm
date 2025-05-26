package storage

import (
	"context"
	"fmt"
	"github.com/StratuStore/fsm/internal/fsm/core"
	"go.mongodb.org/mongo-driver/bson"
)

func (s *DirectoryStorage) WithPagination(
	ctx context.Context,
	filter bson.D,
	offset, limit uint,
	sortByField string,
	sortOrder int,
) (*core.Directory, error) {
	pipeline := []bson.M{
		{"$match": filter},
		{
			"$facet": bson.M{
				"metadata": []bson.M{
					{"$project": bson.M{
						"_id":               1,
						"userID":            1,
						"parentDirectoryID": 1,
						"path":              1,
						"name":              1,
						"createdAt":         1,
						"updatedAt":         1,
						"public":            1,
						"size":              1,
						"directoriesCount":  1,
						"filesCount":        1,
					}},
				},
				"items": []bson.M{
					{"$project": bson.M{
						"allItems": bson.M{
							"$concatArrays": []interface{}{
								bson.M{
									"$map": bson.M{
										"input": bson.M{
											"$sortArray": bson.M{
												"input":  "$directories",
												"sortBy": bson.M{sortByField: sortOrder},
											},
										},
										"in": bson.M{"$mergeObjects": []interface{}{"$$this", bson.M{"_type": "dir"}}},
									},
								},
								bson.M{
									"$map": bson.M{
										"input": bson.M{
											"$sortArray": bson.M{
												"input":  "$files",
												"sortBy": bson.M{sortByField: sortOrder},
											},
										},
										"in": bson.M{"$mergeObjects": []interface{}{"$$this", bson.M{"_type": "file"}}},
									},
								},
							},
						},
					}},
					{"$project": bson.M{
						"items": bson.M{"$slice": []interface{}{"$allItems", int(offset), int(limit)}},
					}},
				},
			},
		},
		{
			"$project": bson.M{
				"_id":               bson.M{"$arrayElemAt": []interface{}{"$metadata._id", 0}},
				"userID":            bson.M{"$arrayElemAt": []interface{}{"$metadata.userID", 0}},
				"parentDirectoryID": bson.M{"$arrayElemAt": []interface{}{"$metadata.parentDirectoryID", 0}},
				"path":              bson.M{"$arrayElemAt": []interface{}{"$metadata.path", 0}},
				"name":              bson.M{"$arrayElemAt": []interface{}{"$metadata.name", 0}},
				"createdAt":         bson.M{"$arrayElemAt": []interface{}{"$metadata.createdAt", 0}},
				"updatedAt":         bson.M{"$arrayElemAt": []interface{}{"$metadata.updatedAt", 0}},
				"public":            bson.M{"$arrayElemAt": []interface{}{"$metadata.public", 0}},
				"size":              bson.M{"$arrayElemAt": []interface{}{"$metadata.size", 0}},
				"directoriesCount":  bson.M{"$arrayElemAt": []interface{}{"$metadata.directoriesCount", 0}},
				"filesCount":        bson.M{"$arrayElemAt": []interface{}{"$metadata.filesCount", 0}},
				"directories": bson.M{
					"$map": bson.M{
						"input": bson.M{
							"$filter": bson.M{
								"input": bson.M{"$arrayElemAt": []interface{}{"$items.items", 0}},
								"cond":  bson.M{"$eq": []string{"$$this._type", "dir"}},
							},
						},
						"in": bson.M{
							"$arrayToObject": bson.M{
								"$filter": bson.M{
									"input": bson.M{"$objectToArray": "$$this"},
									"cond":  bson.M{"$ne": []string{"$$this.k", "_type"}},
								},
							},
						},
					},
				},
				"files": bson.M{
					"$map": bson.M{
						"input": bson.M{
							"$filter": bson.M{
								"input": bson.M{"$arrayElemAt": []interface{}{"$items.items", 0}},
								"cond":  bson.M{"$eq": []string{"$$this._type", "file"}},
							},
						},
						"in": bson.M{
							"$arrayToObject": bson.M{
								"$filter": bson.M{
									"input": bson.M{"$objectToArray": "$$this"},
									"cond":  bson.M{"$ne": []string{"$$this.k", "_type"}},
								},
							},
						},
					},
				},
			},
		},
	}

	cursor, err := s.db.Collection(DirectoryCollection).Aggregate(ctx, pipeline)
	if err != nil {
		return nil, fmt.Errorf("failed to execute aggregation: %w", err)
	}
	defer cursor.Close(ctx)

	var directories []core.Directory
	if err = cursor.All(ctx, &directories); err != nil {
		return nil, fmt.Errorf("failed to decode directory: %w", err)
	}

	if len(directories) == 0 {
		return nil, fmt.Errorf("directory not found")
	}

	return &directories[0], nil
}
