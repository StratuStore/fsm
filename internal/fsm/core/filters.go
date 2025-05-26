package core

import (
	"go.mongodb.org/mongo-driver/bson"
	"time"
)

type Type uint

const (
	Both Type = iota
	FilesOnly
	DirectoriesOnly
)

type Filter struct {
	Type          Type      `query:"type" validate:"-"` // uint: 0 for both (file and directory), 1 for files only, 2 for directories only
	Name          string    `query:"name" validate:"-"`
	CreatedAtFrom time.Time `query:"createdAtFrom" validate:"-"`
	CreatedAtTo   time.Time `query:"createdAtTo" validate:"-"`
	UpdatedAtFrom time.Time `query:"updatedAtFrom" validate:"-"`
	UpdatedAtTo   time.Time `query:"updatedAtTo" validate:"-"`
	Public        *bool     `query:"public" validate:"-"`
	Size          *uint     `query:"size" validate:"-"`
	Starred       *bool     `query:"starred" validate:"-"`
	// Files only
	Extensions []string `query:"extensions" validate:"-"`
}

func (f *Filter) ToMongoFilters() (directoriesFilter bson.D, filesFilter bson.D) {
	if f.Name != "" {
		directoriesFilter = append(directoriesFilter, bson.E{"name", f.Name})
		filesFilter = append(filesFilter, bson.E{"name", f.Name})
	}

	directoriesFilter = append(directoriesFilter,
		bson.E{"createdAt", bson.D{{"$gte", f.CreatedAtFrom}, {"$lte", f.CreatedAtTo}}},
		bson.E{"updatedAt", bson.D{{"$gte", f.UpdatedAtTo}, {"$lte", f.UpdatedAtTo}}},
	)
	filesFilter = append(filesFilter,
		bson.E{"createdAt", bson.D{{"$gte", f.CreatedAtFrom}, {"$lte", f.CreatedAtTo}}},
		bson.E{"updatedAt", bson.D{{"$gte", f.UpdatedAtTo}, {"$lte", f.UpdatedAtTo}}},
	)

	if f.Public != nil {
		directoriesFilter = append(directoriesFilter, bson.E{"public", *f.Public})
		filesFilter = append(filesFilter, bson.E{"public", *f.Public})
	}

	if f.Size != nil {
		directoriesFilter = append(directoriesFilter, bson.E{"size", *f.Size})
		filesFilter = append(filesFilter, bson.E{"size", *f.Size})
	}

	if f.Starred != nil {
		directoriesFilter = append(directoriesFilter, bson.E{"starred", *f.Starred})
		filesFilter = append(filesFilter, bson.E{"starred", *f.Starred})
	}

	if len(f.Extensions) != 0 {
		directoriesFilter = nil
		filesFilter = append(filesFilter, bson.E{"extension", bson.D{{"$in", f.Extensions}}})
	}

	if f.Type == FilesOnly {
		directoriesFilter = nil
	}
	if f.Type == DirectoriesOnly {
		filesFilter = nil
	}

	return directoriesFilter, filesFilter
}
