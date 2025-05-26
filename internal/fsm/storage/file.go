package storage

import (
	"context"
	"fmt"
	"github.com/StratuStore/fsm/internal/fsm/core"
	"github.com/mbretter/go-mongodb/types"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

const FileCollection = "files"

type FileStorage struct {
	Storage
}

func NewFileStorage(s *Storage) *FileStorage {
	return &FileStorage{*s}
}

func (s *FileStorage) GetDirectory(ctx context.Context, id types.ObjectId) (*core.Directory, error) {
	db := s.db

	filter := bson.D{{"_id", id}}

	var directory core.Directory
	err := db.Collection(DirectoryCollection).
		FindOne(ctx, filter).
		Decode(&directory)

	return &directory, err
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

	dir, err := s.GetDirectory(ctx, parentDirID)
	if err != nil {
		return nil, fmt.Errorf("unable to get parent directory: %w", err)
	}

	file := core.File{
		UserID:            userID,
		ParentDirectoryID: string(parentDirID),
		Starred:           false,
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
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
	update := bson.D{{"$push", bson.D{{"files", file}}}, {"$inc", bson.D{{"filesCount", 1}, {"size", size}}}}
	_, err = db.Collection(DirectoryCollection).
		UpdateOne(
			ctx,
			filter,
			update,
		)
	if err != nil {
		return nil, fmt.Errorf("unable to insert file: %w", err)
	}

	dirIDs := make([]types.ObjectId, 0, len(dir.Path))
	for _, d := range dir.Path {
		dirIDs = append(dirIDs, d.ID)
	}

	filter = bson.D{{"_id", bson.D{{"$in", dirIDs}}}}
	update = bson.D{{"$inc", bson.D{{"size", size}}}}
	_, err = db.Collection(DirectoryCollection).
		UpdateMany(
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
	dir, err := s.GetDirectory(ctx, types.ObjectId(file.ParentDirectoryID))
	if err != nil {
		return fmt.Errorf("unable to get parent directory: %w", err)
	}

	filter := bson.D{{"_id", types.ObjectId(file.ParentDirectoryID)}}
	update := bson.D{
		{"$pull", bson.D{{"files", bson.D{{"_id", id}}}}},
		{"$inc", bson.D{{"filesCount", -1}, {"size", -file.Size}}},
	}
	_, err = db.Collection(DirectoryCollection).
		UpdateOne(
			ctx,
			filter,
			update,
		)
	if err != nil {
		return fmt.Errorf("unable to delete file from parent: %w", err)
	}

	dirIDs := make([]types.ObjectId, 0, len(dir.Path))
	for _, d := range dir.Path {
		dirIDs = append(dirIDs, d.ID)
	}

	filter = bson.D{{"_id", bson.D{{"$in", dirIDs}}}}
	update = bson.D{{"$inc", bson.D{{"size", -file.Size}}}}
	_, err = db.Collection(DirectoryCollection).
		UpdateMany(
			ctx,
			filter,
			update,
		)

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

	timestamp := time.Now()

	filter := bson.D{{"_id", id}}
	update := bson.D{{"$set", bson.D{{"name", newName}, {"updatedAt", timestamp}}}}
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
	update = bson.D{{"$set", bson.D{{"files.$.name", newName}, {"files.$.updatedAt", timestamp}}}}
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
	file.UpdatedAt = time.Now()
	fromDir, err := s.GetDirectory(ctx, types.ObjectId(file.ParentDirectoryID))
	if err != nil {
		return fmt.Errorf("unable to get parent directory: %w", err)
	}
	toDir, err := s.GetDirectory(ctx, toID)
	if err != nil {
		return fmt.Errorf("unable to get target directory: %w", err)
	}

	filter := bson.D{{"_id", types.ObjectId(file.ParentDirectoryID)}}
	update := bson.D{
		{"$pull", bson.D{{"files", bson.D{{"_id", id}}}}},
		{"$inc", bson.D{{"filesCount", -1}, {"size", -file.Size}}},
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

	filter = bson.D{{"_id", id}}
	update = bson.D{{"$set", bson.D{{"parentDirectoryID", string(toID)}, {"updatedAt", file.UpdatedAt}}}}
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
	update = bson.D{
		{"$push", bson.D{{"files", file}}},
		{"$inc", bson.D{{"filesCount", 1}, {"size", file.Size}}},
	}
	_, err = db.Collection(DirectoryCollection).
		UpdateOne(
			ctx,
			filter,
			update,
		)

	fromDirIDs := make([]types.ObjectId, 0, len(fromDir.Path))
	for _, d := range fromDir.Path {
		fromDirIDs = append(fromDirIDs, d.ID)
	}
	toDirIDs := make([]types.ObjectId, 0, len(toDir.Path))
	for _, d := range toDir.Path {
		toDirIDs = append(toDirIDs, d.ID)
	}

	filter = bson.D{{"_id", bson.D{{"$in", fromDirIDs}}}}
	update = bson.D{{"$inc", bson.D{{"size", -file.Size}}}}
	_, err = db.Collection(DirectoryCollection).
		UpdateMany(
			ctx,
			filter,
			update,
		)
	if err != nil {
		return fmt.Errorf("unable to update directories sizes: %w", err)
	}

	filter = bson.D{{"_id", bson.D{{"$in", toDirIDs}}}}
	update = bson.D{{"$inc", bson.D{{"size", file.Size}}}}
	_, err = db.Collection(DirectoryCollection).
		UpdateMany(
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
	file.UpdatedAt = time.Now()
	dir, err := s.GetDirectory(ctx, types.ObjectId(file.ParentDirectoryID))
	if err != nil {
		return fmt.Errorf("unable to get directory: %w", err)
	}

	filter := bson.D{{"_id", id}}
	update := bson.D{{"$set", bson.D{{"size", size}, {"updatedAt", file.UpdatedAt}}}}
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
	update = bson.D{{"$set", bson.D{{"files.$.size", size}, {"files.$.updatedAt", file.UpdatedAt}}}}
	_, err = db.Collection(DirectoryCollection).
		UpdateMany(
			ctx,
			filter,
			update,
		)
	if err != nil {
		return fmt.Errorf("unable to update file size inside dir: %w", err)
	}

	dirIDs := make([]types.ObjectId, 0, len(dir.Path))
	for _, d := range dir.Path {
		dirIDs = append(dirIDs, d.ID)
	}

	filter = bson.D{{"_id", bson.D{{"$in", dirIDs}}}}
	update = bson.D{{"$inc", bson.D{{"size", diff}}}}
	_, err = db.Collection(DirectoryCollection).
		UpdateMany(
			ctx,
			filter,
			update,
		)

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
	timestamp := time.Now()

	filter := bson.D{{"_id", id}}
	update := bson.D{{"$set", bson.D{{field, value}, {"updatedAt", timestamp}}}}
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
	update = bson.D{{"$set", bson.D{{"files.$." + field, value}, {"files.$.updatedAt", timestamp}}}}
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
