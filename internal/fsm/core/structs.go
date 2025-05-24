package core

import (
	"github.com/mbretter/go-mongodb/types"
	"time"
)

type PathElement struct {
	ID   types.ObjectId `bson:"_id" json:"id"`
	Name string         `bson:"name"`
}

type Directory struct {
	ID                types.ObjectId `json:"id" bson:"_id,omitempty"`
	UserID            string         `json:"userID" bson:"userID"`
	ParentDirectoryID string         `bson:"parentDirectoryID" json:"parentDirectoryID,omitempty"`
	Path              []PathElement  `json:"path" bson:"path"`
	Name              string         `json:"name" bson:"name"`
	CreatedAt         time.Time      `json:"createdAt" bson:"createdAt,omitempty"`
	UpdatedAt         time.Time      `json:"updatedAt" bson:"updatedAt,omitempty"`
	Public            bool           `json:"public" bson:"public"`
	DirectoriesCount  uint           `json:"directoriesCount" bson:"directoriesCount"`
	Directories       []Directory    `json:"directories" bson:"directories"`
	FilesCount        uint           `json:"filesCount" bson:"filesCount"`
	Files             []File         `json:"files" bson:"files"`
	Size              uint           `json:"size" bson:"size"`
}

type File struct {
	ID        types.ObjectId    `json:"id" bson:"_id,omitempty"`
	UserID    string            `json:"userID" bson:"userID"`
	CreatedAt time.Time         `json:"createdAt" bson:"createdAt"`
	UpdatedAt time.Time         `json:"updatedAt" bson:"updatedAt"`
	Starred   bool              `json:"starred" bson:"starred"`
	Size      uint              `bson:"size" json:"size"`
	Public    bool              `json:"public" bson:"public"`
	Name      string            `json:"name" bson:"name"`
	Extension string            `json:"extension" bson:"extension"`
	Attrs     map[string]string `json:"attrs" bson:"attrs"`
}
