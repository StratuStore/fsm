package storage

import (
	"context"
	"fmt"
	"github.com/StratuStore/fsm/internal/fsm/core"
	"github.com/mbretter/go-mongodb/types"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
	"slices"
	"time"
)

const DirectoryCollection = "directories"

type DirectoryStorage struct {
	Storage
}

func NewDirectoryStorage(s *Storage) *DirectoryStorage {
	return &DirectoryStorage{*s}
}

func (s *DirectoryStorage) Get(ctx context.Context, id types.ObjectId) (*core.Directory, error) {
	db := s.db

	filter := bson.D{{"_id", id}}

	var directory core.Directory
	err := db.Collection(DirectoryCollection).
		FindOne(ctx, filter).
		Decode(&directory)

	return &directory, err
}

func (s *DirectoryStorage) GetWithPagination(
	ctx context.Context,
	id types.ObjectId,
	offset, limit uint,
	sortByField string,
	sortOrder int,
) (*core.Directory, error) {
	filter := bson.D{{"_id", id}}

	return s.WithPagination(ctx, filter, offset, limit, sortByField, sortOrder)
}

func (s *DirectoryStorage) GetRoot(
	ctx context.Context,
	userID string,
	offset, limit uint,
	sortByField string,
	sortOrder int,
) (*core.Directory, error) {
	filter := bson.D{{"userID", userID}, {"path", nil}}

	if err := s.db.Collection(DirectoryCollection).FindOne(ctx, filter).Err(); err != nil {
		return nil, err
	}

	return s.WithPagination(ctx, filter, offset, limit, sortByField, sortOrder)
}

func (s *DirectoryStorage) CreateRoot(
	ctx context.Context,
	userID string,
	offset, limit uint,
	sortByField string,
	sortOrder int,
) (*core.Directory, error) {
	db := s.db

	directory := core.Directory{
		UserID:           userID,
		Path:             nil,
		Name:             "root",
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
		Public:           false,
		DirectoriesCount: 0,
		Directories:      []core.Directory{},
		FilesCount:       0,
		Files:            []core.File{},
		Size:             0,
	}

	result, err := db.Collection(DirectoryCollection).
		InsertOne(ctx, directory)
	if err != nil {
		return nil, fmt.Errorf("unable to insert root folder: %w", err)
	}

	id, ok := result.InsertedID.(primitive.ObjectID)
	if !ok {
		return nil, fmt.Errorf("unable to contert id to string: %w", err)
	}

	directory.ID = types.ObjectId(id.Hex())

	filter := bson.D{{"userID", userID}, {"path", nil}}

	return s.WithPagination(ctx, filter, offset, limit, sortByField, sortOrder)
}

func (s *DirectoryStorage) Create(ctx context.Context, parentDirID types.ObjectId, userID string, name string) (*core.Directory, error) {
	db := s.db

	parentDir, err := s.Get(ctx, parentDirID)
	if err != nil {
		return nil, fmt.Errorf("unable to find parentDir: %w", err)
	}
	path := slices.Clone(parentDir.Path)
	path = append(path, core.PathElement{parentDirID, parentDir.Name})

	directory := core.Directory{
		UserID:            userID,
		Path:              path,
		ParentDirectoryID: string(parentDirID),
		Name:              name,
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
		Public:            false,
		DirectoriesCount:  0,
		Directories:       []core.Directory{},
		FilesCount:        0,
		Files:             []core.File{},
		Size:              0,
	}

	result, err := db.Collection(DirectoryCollection).
		InsertOne(ctx, directory)
	if err != nil {
		return nil, fmt.Errorf("unable to insert root folder: %w", err)
	}

	id, ok := result.InsertedID.(primitive.ObjectID)
	if !ok {
		return nil, fmt.Errorf("unable to contert id to string: %w", err)
	}
	directory.ID = types.ObjectId(id.Hex())

	directory.Directories = nil
	directory.Files = nil
	filter := bson.D{{"_id", parentDirID}}
	update := bson.D{{"$push", bson.D{{"directories", directory}}}, {"$inc", bson.D{{"directoriesCount", 1}}}}
	_, err = db.Collection(DirectoryCollection).
		UpdateOne(
			ctx,
			filter,
			update,
		)
	directory.Directories = []core.Directory{}
	directory.Files = []core.File{}

	return &directory, err
}

func (s *DirectoryStorage) Delete(ctx context.Context, id types.ObjectId) error {
	db := s.db

	dir, err := s.Get(ctx, id)
	if err != nil {
		return fmt.Errorf("unable to find dir: %w", err)
	}

	filter := bson.D{{"_id", types.ObjectId(dir.ParentDirectoryID)}}
	update := bson.D{
		{"$pull", bson.D{{"directories", bson.D{{"_id", id}}}}},
		{"$inc", bson.D{{"directoriesCount", -1}}}}
	_, err = db.Collection(DirectoryCollection).
		UpdateOne(
			ctx,
			filter,
			update,
		)
	if err != nil {
		return fmt.Errorf("unable to delete dir from parent: %w", err)
	}

	if err := IncrementSizes(db, ctx, dir.Path, -int(dir.Size)); err != nil {
		return err
	}

	return s.StupidDelete(ctx, id)
}

func (s *DirectoryStorage) StupidDelete(ctx context.Context, id types.ObjectId) error {
	db := s.db

	filter := bson.D{{"_id", id}}
	_, err := db.Collection(DirectoryCollection).DeleteOne(ctx, filter)

	return err
}

func (s *FileStorage) StupidDeleteFile(ctx context.Context, id types.ObjectId) error {
	db := s.db

	filter := bson.D{{"_id", id}}
	_, err := db.Collection(FileCollection).DeleteOne(ctx, filter)

	return err
}

func (s *DirectoryStorage) Rename(ctx context.Context, id types.ObjectId, newName string) error {
	db := s.db
	timestamp := time.Now()

	filter := bson.D{{"_id", id}}
	update := bson.D{{"$set", bson.D{{"name", newName}, {"updatedAt", timestamp}}}}
	_, err := db.Collection(DirectoryCollection).
		UpdateOne(
			ctx,
			filter,
			update,
		)
	if err != nil {
		return fmt.Errorf("unable to update name of the directory: %w", err)
	}

	filter = bson.D{{"directories._id", id}}
	update = bson.D{{"$set", bson.D{{"directories.$.name", newName}, {"directories.$.updatedAt", timestamp}}}}
	_, err = db.Collection(DirectoryCollection).
		UpdateMany(
			ctx,
			filter,
			update,
		)
	if err != nil {
		return fmt.Errorf("unable to update name of the directory: %w", err)
	}

	filter = bson.D{{"path._id", id}}
	update = bson.D{{"$set", bson.D{{"path.$.name", newName}}}}
	_, err = db.Collection(DirectoryCollection).
		UpdateMany(
			ctx,
			filter,
			update,
		)

	arrayFilter := []any{bson.D{{"num._id", id}}}
	update = bson.D{{"$set", bson.D{{"directories.$[].path.$[num].name", newName}}}}
	_, err = db.Collection(DirectoryCollection).
		UpdateMany(
			ctx,
			bson.D{},
			update,
			options.Update().SetArrayFilters(options.ArrayFilters{Filters: arrayFilter}),
		)

	return err
}

func (s *DirectoryStorage) Move(ctx context.Context, id, toID types.ObjectId) error {
	db := s.db
	timestamp := time.Now()

	dir, err := s.Get(ctx, id)
	if err != nil {
		return fmt.Errorf("unable to find dir: %w", err)
	}
	dir.UpdatedAt = timestamp
	fromDir, err := s.Get(ctx, types.ObjectId(dir.ParentDirectoryID))
	if err != nil {
		return fmt.Errorf("unable to find initial dir: %w", err)
	}
	toDir, err := s.Get(ctx, toID)
	if err != nil {
		return fmt.Errorf("unable to find target dir: %w", err)
	}

	filter := bson.D{{"_id", types.ObjectId(dir.ParentDirectoryID)}}
	update := bson.D{
		{"$pull", bson.D{{"directories", bson.D{{"_id", id}}}}},
		{"$inc", bson.D{{"directoriesCount", -1}, {"size", -int(dir.Size)}}},
	}
	_, err = db.Collection(DirectoryCollection).
		UpdateOne(
			ctx,
			filter,
			update,
		)
	if err != nil {
		return fmt.Errorf("unable to delete dir from parent: %w", err)
	}
	if err := UpdateEmbeddedSize(db, ctx, types.ObjectId(dir.ParentDirectoryID), -int(dir.Size)); err != nil {
		return err
	}

	path := slices.Clone(toDir.Path)
	path = append(path, core.PathElement{toDir.ID, toDir.Name})

	filter = bson.D{{"_id", id}}
	update = bson.D{{"$set", bson.D{{"parentDirectoryID", string(toID)}, {"path", path}, {"updatedAt", timestamp}}}}
	_, err = db.Collection(DirectoryCollection).
		UpdateMany(
			ctx,
			filter,
			update,
		)
	if err != nil {
		return fmt.Errorf("unable to update dir parentID: %w", err)
	}

	oldPath := slices.Clone(dir.Path)
	dir.Directories = nil
	dir.Files = nil
	dir.Path = path
	dir.ParentDirectoryID = string(toID)
	filter = bson.D{{"_id", toID}}
	update = bson.D{
		{"$push", bson.D{{"directories", dir}}},
		{"$inc", bson.D{{"directoriesCount", 1}, {"size", dir.Size}}},
	}
	_, err = db.Collection(DirectoryCollection).
		UpdateOne(
			ctx,
			filter,
			update,
		)
	if err != nil {
		return fmt.Errorf("unable to update toDir: %w", err)
	}
	if err := UpdateEmbeddedSize(db, ctx, toID, int(dir.Size)); err != nil {
		return err
	}

	if err := s.UpdatePath(ctx, id, oldPath, dir.Path); err != nil {
		return err
	}

	if err := IncrementSizes(db, ctx, fromDir.Path, -int(dir.Size)); err != nil {
		return err
	}
	if err := IncrementSizes(db, ctx, toDir.Path, int(dir.Size)); err != nil {
		return err
	}

	return err
}

func (s *DirectoryStorage) Star(ctx context.Context, id types.ObjectId) error {
	directory, err := s.Get(ctx, id)
	if err != nil {
		return fmt.Errorf("unable to find directory: %w", err)
	}

	return s.UpdateField(ctx, id, "starred", !directory.Starred)
}

func (s *DirectoryStorage) Share(ctx context.Context, id types.ObjectId, mode bool) error {
	return s.UpdateField(ctx, id, "public", mode)
}

func (s *DirectoryStorage) UpdateField(ctx context.Context, id types.ObjectId, field string, value any) error {
	db := s.db
	timestamp := time.Now()

	filter := bson.D{{"_id", id}}
	update := bson.D{{"$set", bson.D{{field, value}, {"updatedAt", timestamp}}}}
	_, err := db.Collection(DirectoryCollection).
		UpdateMany(
			ctx,
			filter,
			update,
		)
	if err != nil {
		return fmt.Errorf("unable to update file field: %w", err)
	}

	filter = bson.D{{"directories._id", id}}
	update = bson.D{{"$set", bson.D{{"directories.$." + field, value}, {"directories.$.updatedAt", timestamp}}}}
	_, err = db.Collection(DirectoryCollection).
		UpdateMany(
			ctx,
			filter,
			update,
		)
	if err != nil {
		return fmt.Errorf("unable to update file field inside dir: %w", err)
	}

	return err
}

func (s *DirectoryStorage) UpdatePath(ctx context.Context, id types.ObjectId, oldPath []core.PathElement, newPath []core.PathElement) error {
	db := s.db

	oldPathIDs := make([]types.ObjectId, len(oldPath))
	for num, p := range oldPath {
		oldPathIDs[num] = p.ID
	}

	filter := bson.D{{"path._id", id}}
	update := bson.D{
		{"$pull", bson.D{{"path", bson.D{{"_id", bson.D{{"$in", oldPathIDs}}}}}}},
	}
	_, err := db.Collection(DirectoryCollection).
		UpdateMany(
			ctx,
			filter,
			update,
		)
	if err != nil {
		return fmt.Errorf("unable to update path: %w", err)
	}
	update = bson.D{
		{"$push", bson.D{{"path", bson.D{{"$each", newPath}, {"$position", 0}}}}},
	}
	_, err = db.Collection(DirectoryCollection).
		UpdateMany(
			ctx,
			filter,
			update,
		)
	if err != nil {
		return fmt.Errorf("unable to update path2: %w", err)
	}

	arrayFilter := []any{bson.D{{"num.path", bson.D{{"$elemMatch", bson.D{{"_id", id}}}}}}}
	filter = bson.D{}
	update = bson.D{
		{"$pull", bson.D{{"directories.$[num].path", bson.D{{"_id", bson.D{{"$in", oldPathIDs}}}}}}},
	}
	_, err = db.Collection(DirectoryCollection).
		UpdateMany(
			ctx,
			filter,
			update,
			options.Update().SetArrayFilters(options.ArrayFilters{Filters: arrayFilter}),
		)
	if err != nil {
		return fmt.Errorf("unable to update embedded path: %w", err)
	}
	update = bson.D{
		{"$push", bson.D{{"directories.$[num].path", bson.D{{"$each", newPath}, {"$position", 0}}}}},
	}
	_, err = db.Collection(DirectoryCollection).
		UpdateMany(
			ctx,
			filter,
			update,
			options.Update().SetArrayFilters(options.ArrayFilters{Filters: arrayFilter}),
		)
	if err != nil {
		return fmt.Errorf("unable to update embedded path2: %w", err)
	}

	return nil
}
