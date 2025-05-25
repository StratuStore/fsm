package storage

import (
	"context"
	"fmt"
	"github.com/StratuStore/fsm/internal/fsm/core"
	"github.com/mbretter/go-mongodb/types"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

const FileCollection = "files"

type FileStorage struct {
	Storage
}

func NewFileStorage(s *Storage) *FileStorage {
	return &FileStorage{*s}
}

func (s *FileStorage) Get(ctx context.Context, id types.ObjectId) (*core.File, error) {
	db := s.db

	filter := bson.D{{"_id", id}}

	var file core.File
	err := db.Collection(FileCollection).
		FindOne(ctx, filter).
		Decode(&file)

	return &file, err
}

func (s *FileStorage) Create(ctx context.Context, parentDirID types.ObjectId, userID string, name, extension string, size uint) (*core.File, error) {
	db := s.db

	file := core.File{
		UserID:            userID,
		ParentDirectoryID: string(parentDirID),
		Starred:           false,
		Size:              size,
		Public:            false,
		Name:              name,
		Extension:         extension,
		Attrs:             map[string]string{},
	}

	result, err := db.Collection(FileCollection).
		InsertOne(ctx, file)
	if err != nil {
		return nil, fmt.Errorf("unable to insert file: %w", err)
	}

	id, ok := result.InsertedID.(primitive.ObjectID)
	if !ok {
		return nil, fmt.Errorf("unable to contert id to string: %w", err)
	}
	file.ID = types.ObjectId(id.Hex())

	filter := bson.D{{"_id", parentDirID}}
	update := bson.D{{"$push", bson.D{{"files", file}}}, {"$inc", bson.D{{"filesCount", 1}}}}
	_, err = db.Collection(DirectoryCollection).
		UpdateOne(
			ctx,
			filter,
			update,
		)

	return &file, err
}

func (s *FileStorage) Delete(ctx context.Context, id types.ObjectId) error {
	db := s.db

	file, err := s.Get(ctx, id)
	if err != nil {
		return fmt.Errorf("unable to find file: %w", err)
	}

	filter := bson.D{{"_id", types.ObjectId(file.ParentDirectoryID)}}
	update := bson.D{{"$pull", bson.D{{"files", bson.D{{"_id", id}}}}}, {"$inc", bson.D{{"filesCount", -1}}}}
	_, err = db.Collection(DirectoryCollection).
		UpdateOne(
			ctx,
			filter,
			update,
		)
	if err != nil {
		return fmt.Errorf("unable to delete file from parent: %w", err)
	}

	return s.StupidDelete(ctx, id)
}

func (s *FileStorage) StupidDelete(ctx context.Context, id types.ObjectId) error {
	db := s.db

	filter := bson.D{{"_id", id}}
	_, err := db.Collection(FileCollection).DeleteOne(ctx, filter)

	return err
}

func (s *FileStorage) Rename(ctx context.Context, id types.ObjectId, newName string) error {
	db := s.db

	filter := bson.D{{"_id", id}}
	update := bson.D{{"$set", bson.D{{"name", newName}}}}
	_, err := db.Collection(FileCollection).
		UpdateOne(
			ctx,
			filter,
			update,
		)
	if err != nil {
		return fmt.Errorf("unable to update name of the directory: %w", err)
	}

	filter = bson.D{{"files._id", id}}
	update = bson.D{{"$set", bson.D{{"files.$.name", newName}}}}
	_, err = db.Collection(DirectoryCollection).
		UpdateMany(
			ctx,
			filter,
			update,
		)

	return err
}

func (s *FileStorage) Move(ctx context.Context, id, toID types.ObjectId) error {
	db := s.db

	file, err := s.Get(ctx, id)
	if err != nil {
		return fmt.Errorf("unable to find file: %w", err)
	}

	filter := bson.D{{"_id", types.ObjectId(file.ParentDirectoryID)}}
	update := bson.D{{"$pull", bson.D{{"files", bson.D{{"_id", id}}}}}, {"$inc", bson.D{{"filesCount", -1}}}}
	_, err = db.Collection(DirectoryCollection).
		UpdateOne(
			ctx,
			filter,
			update,
		)
	if err != nil {
		return fmt.Errorf("unable to delete dir from parent: %w", err)
	}

	filter = bson.D{{"_id", id}}
	update = bson.D{{"$set", bson.D{{"parentDirectoryID", string(toID)}}}}
	_, err = db.Collection(FileCollection).
		UpdateMany(
			ctx,
			filter,
			update,
		)
	if err != nil {
		return fmt.Errorf("unable to update file parentID: %w", err)
	}

	filter = bson.D{{"_id", toID}}
	update = bson.D{{"$push", bson.D{{"files", file}}}, {"$inc", bson.D{{"filesCount", 1}}}}
	_, err = db.Collection(DirectoryCollection).
		UpdateOne(
			ctx,
			filter,
			update,
		)

	return err
}

func (s *FileStorage) Update(ctx context.Context, id types.ObjectId, size uint) error {
	db := s.db

	file, err := s.Get(ctx, id)
	if err != nil {
		return fmt.Errorf("unable to find file: %w", err)
	}
	diff := size - file.Size

	filter := bson.D{{"_id", id}}
	update := bson.D{{"$set", bson.D{{"size", size}}}}
	_, err = db.Collection(FileCollection).
		UpdateMany(
			ctx,
			filter,
			update,
		)
	if err != nil {
		return fmt.Errorf("unable to update file size: %w", err)
	}

	filter = bson.D{{"files._id", id}}
	update = bson.D{{"$set", bson.D{{"files.$.size", size}}}}
	_, err = db.Collection(DirectoryCollection).
		UpdateMany(
			ctx,
			filter,
			update,
		)
	if err != nil {
		return fmt.Errorf("unable to update file size inside dir: %w", err)
	}

	parentDirectoryID := file.ParentDirectoryID
	for parentDirectoryID != "" {
		var dir core.Directory

		filter = bson.D{{"_id", parentDirectoryID}}
		update = bson.D{{"$inc", bson.D{{"size", diff}}}}
		err = db.Collection(DirectoryCollection).
			FindOneAndUpdate(
				ctx,
				filter,
				update,
			).
			Decode(&dir)
		if err != nil {
			return fmt.Errorf("unable to update size of dir: %w", err)
		}
		parentDirectoryID = dir.ParentDirectoryID
	}

	return err
}

func (s *FileStorage) Star(ctx context.Context, id types.ObjectId) error {
	file, err := s.Get(ctx, id)
	if err != nil {
		return fmt.Errorf("unable to find file: %w", err)
	}

	return s.UpdateField(ctx, id, "starred", !file.Starred)
}

func (s *FileStorage) Share(ctx context.Context, id types.ObjectId, mode bool) error {
	return s.UpdateField(ctx, id, "public", mode)
}

func (s *FileStorage) UpdateField(ctx context.Context, id types.ObjectId, field string, value any) error {
	db := s.db

	filter := bson.D{{"_id", id}}
	update := bson.D{{"$set", bson.D{{field, value}}}}
	_, err := db.Collection(FileCollection).
		UpdateMany(
			ctx,
			filter,
			update,
		)
	if err != nil {
		return fmt.Errorf("unable to update file field: %w", err)
	}

	filter = bson.D{{"files._id", id}}
	update = bson.D{{"$set", bson.D{{"files.$." + field, value}}}}
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
