package endpoint

import (
	"context"
	"mime/multipart"

	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/log"
	"github.com/wvoliveira/burnit/internal/pkg/service"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Set struct {
	CreateEndpoint endpoint.Endpoint
}

func New(svc service.Service, logger log.Logger) Set {
	var createEndpoint endpoint.Endpoint
	{
		createEndpoint = MakeCreateEndpoint(svc)
		createEndpoint = LoggingMiddleware(log.With(logger, "method", "Create"))(createEndpoint)
	}

	return Set{
		CreateEndpoint: createEndpoint,
	}
}

// Concat implements the service interface, so Set may be used as a
// service. This is primarily useful in the context of a client library.
func (s Set) Create(ctx context.Context, fileName string, file multipart.File, text []byte) (string, error) {
	resp, err := s.CreateEndpoint(ctx, CreateRequest{
		FileName: fileName,
		File:     file,
		Text:     text,
	})
	if err != nil {
		return "", err
	}
	response := resp.(CreateResponse)
	return response.V, response.Err
}

// MakeCreateEndpoint constructs a Create endpoint wrapping the service.
func MakeCreateEndpoint(s service.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(CreateRequest)
		v, err := s.Create(ctx, req.FileName, req.File, req.Text)
		return CreateResponse{V: v, Err: err}, nil
	}
}

// compile time assertions for our response types implementing endpoint.Failer.
var (
	_ endpoint.Failer = CreateResponse{}
)

// CreateRequest collects the request parameters for the Create method.
type CreateRequest struct {
	ID       primitive.ObjectID
	Key      string
	Text     []byte
	FileName string
	File     multipart.File
}

// CreateResponse collects the response values for the Create method.
type CreateResponse struct {
	V   string `json:"v"`
	Err error  `json:"-"` // should be intercepted by Failed/errorEncoder
}

// Failed implements endpoint.Failer.
func (r CreateResponse) Failed() error { return r.Err }
