package storage

import (
	"context"
	"fmt"
	"github.com/StratuStore/fsm/internal/fsm/core"
	"github.com/mbretter/go-mongodb/types"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func IncrementSizes(db *mongo.Database, ctx context.Context, path []core.PathElement, size int) error {
	toDirIDs := make([]types.ObjectId, 0, len(path))
	for _, d := range path {
		toDirIDs = append(toDirIDs, d.ID)
	}

	filter := bson.D{{"_id", bson.D{{"$in", toDirIDs}}}}
	update := bson.D{{"$inc", bson.D{{"size", size}}}}
	_, err := db.Collection(DirectoryCollection).
		UpdateMany(
			ctx,
			filter,
			update,
		)
	if err != nil {
		return fmt.Errorf("unable to update directories sizes: %w", err)
	}

	filter = bson.D{{"directories._id", bson.D{{"$in", toDirIDs}}}}
	update = bson.D{{"$inc", bson.D{{"directories.$.size", size}}}}
	_, err = db.Collection(DirectoryCollection).
		UpdateMany(
			ctx,
			filter,
			update,
		)
	if err != nil {
		return fmt.Errorf("unable to update embedded directories sizes: %w", err)
	}

	return nil
}

func UpdateEmbeddedSize(db *mongo.Database, ctx context.Context, id types.ObjectId, size int) error {
	filter := bson.D{{"directories._id", id}}
	update := bson.D{{"$inc", bson.D{{"directories.$.size", size}}}}
	_, err := db.Collection(DirectoryCollection).
		UpdateMany(
			ctx,
			filter,
			update,
		)
	if err != nil {
		return fmt.Errorf("unable to update embedded directories sizes: %w", err)
	}

	return nil
}
