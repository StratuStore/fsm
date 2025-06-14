package communicator

import (
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/mbretter/go-mongodb/types"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type RequestType int

const (
	CreateType RequestType = iota
	UpdateType
	OpenType
	DeleteType
)

type Request struct {
	ID     uuid.UUID
	Host   string
	Type   RequestType
	FileID uuid.UUID
	Size   uint
}

func NewRequest(requestType RequestType, host string, fileID types.ObjectId, size uint) (*Request, error) {
	objID, err := primitive.ObjectIDFromHex(string(fileID))
	if err != nil {
		return nil, fmt.Errorf("unable to convert fileID to objectID")
	}

	var fileUUID uuid.UUID
	copy(fileUUID[:6], objID[:6])
	copy(fileUUID[10:], objID[6:])
	fileUUID[6] &= 0x0f /* clear version        */
	fileUUID[6] |= 0x40 /* set to version 4     */
	fileUUID[8] &= 0x3f /* clear variant        */
	fileUUID[8] |= 0x80 /* set to IETF variant  */

	return &Request{
		ID:     uuid.New(),
		Type:   requestType,
		Host:   host,
		FileID: fileUUID,
		Size:   size,
	}, nil
}

type Response struct {
	ID           uuid.UUID // must be equal to request.ID
	Host         string
	ConnectionID uuid.UUID
	Err          string
}

func (r *Response) ToReturn() (string, string, error) {
	var err error
	if r.Err != "" {
		err = errors.New(r.Err)
	}

	return r.Host, r.ConnectionID.String(), err
}
