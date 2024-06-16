package service

import (
	"context"
	"mime/multipart"

	"github.com/go-kit/log"
)

// Middleware describes a service (as opposed to endpoint) middleware.
type Middleware func(Service) Service

// LoggingMiddleware takes a logger as a dependency
// and returns a service Middleware.
func LoggingMiddleware(logger log.Logger) Middleware {
	return func(next Service) Service {
		return loggingMiddleware{logger, next}
	}
}

type loggingMiddleware struct {
	logger log.Logger
	next   Service
}

func (mw loggingMiddleware) Create(ctx context.Context, fileName string, file multipart.File, text []byte) (key string, err error) {
	defer func() {
		mw.logger.Log("method", "Create", "fileName", fileName, "file", "-", "text", "-", "err", err)
	}()
	return mw.next.Create(ctx, fileName, file, text)
}
