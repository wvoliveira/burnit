package grpc

import (
	"github.com/go-kit/log"
	"github.com/wvoliveira/burnit/internal/pkg/endpoint"
	"github.com/wvoliveira/burnit/pb"
)

func NewGRPCServer(endpoints endpoint.Set, logger log.Logger) pb.BurnItServer {
	return nil
}
