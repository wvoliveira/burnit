package endpoint

import (
	"context"
	"log"

	"github.com/wvoliveira/burnit/internal/pkg/model"
	"github.com/wvoliveira/burnit/internal/pkg/service"
)

type Endpoint func(ctx context.Context, request interface{}) (response interface{}, err error)

type Middleware func(Endpoint) Endpoint

type Set struct {
	CreateEndpoint Endpoint
}

func New(svc service.Service, logger *log.Logger) Set {
	var createEndpoint Endpoint
	return Set{
		CreateEndpoint: createEndpoint,
	}
}

func MakeCreateEndpoint(s service.Service) Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(model.RequestBody)
		v, err := s.Create(ctx, req.FileName, req.File, req.Text)
		return CreateResponse{V: v, Err: err}, nil
	}
}

type CreateResponse struct {
	V   string `json:"v"`
	Err error  `json:"-"`
}
