package core

import (
	"time"
)

type Directory struct {
	ID          string      `json:"id" bson:"_id"`
	UserID      string      `json:"userID" bson:"userID"`
	Path        []string    `json:"path" bson:"path"`
	CreatedAt   time.Time   `json:"createdAt" bson:"createdAt"`
	UpdatedAt   time.Time   `json:"updatedAt" bson:"updatedAt"`
	Public      bool        `json:"public" bson:"public"`
	Directories []Directory `json:"directories" bson:"directories"`
	Files       []File      `json:"files" bson:"files"`
}

type File struct {
	ID        string            `json:"id" bson:"_id"`
	UserID    string            `json:"userID" bson:"userID"`
	CreatedAt time.Time         `json:"createdAt" bson:"createdAt"`
	UpdatedAt time.Time         `json:"updatedAt" bson:"updatedAt"`
	Public    bool              `json:"public" bson:"public"`
	Name      string            `json:"name" bson:"name"`
	Attrs     map[string]string `json:"attrs" bson:"attrs"`
	Server    string            `json:"server" bson:"server"`
}
