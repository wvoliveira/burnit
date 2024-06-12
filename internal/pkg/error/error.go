package error

import "errors"

var (
	ErrMongoDBNoDocuments = errors.New("not found")
	ErrInternalServer     = errors.New("internal server error")
)
