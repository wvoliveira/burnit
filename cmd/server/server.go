package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"

	"github.com/wvoliveira/burnit/internal/pkg/config"
	"github.com/wvoliveira/burnit/internal/pkg/database"
	"github.com/wvoliveira/burnit/internal/pkg/endpoint"
	"github.com/wvoliveira/burnit/internal/pkg/service"
	transportgrpc "github.com/wvoliveira/burnit/internal/pkg/transport/grpc"
	transporthttp "github.com/wvoliveira/burnit/internal/pkg/transport/http"
	"github.com/wvoliveira/burnit/pb"
	"google.golang.org/grpc"
)

func init() {
	if config.MongoDBURI == "" {
		log.Fatal("You must set your 'MONGODB_URI' environment variable. " +
			"See\n\t https://www.mongodb.com/docs/drivers/go/current/usage-examples/#environment-variable")
	}
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), config.MongoDBTimeout)
	defer cancel()

	kvClient, err := database.NewMongoDB(ctx, config.MongoDBURI)
	if err != nil {
		log.Printf("Error to connect to MongoDB: %s\n", err.Error())
		os.Exit(1)
	}

	defer func() {
		if err := kvClient.Disconnect(ctx); err != nil {
			log.Printf("Error to disconnect from MongoDB: %s\n", err.Error())
			os.Exit(1)
		}
	}()

	var logger log.Logger

	// Build the layers of the service "onion" from the inside out. First, the
	// business logic service; then, the set of endpoints that wrap the service;
	// and finally, a series of concrete transport adapters. The adapters, like
	// the HTTP handler or the gRPC server, are the bridge between Go kit and
	// the interfaces that the transports expect. Note that we're not binding
	// them to ports or anything yet; we'll do that next.
	// Copy from: https://github.com/go-kit/examples/blob/master/addsvc/cmd/addsvc/addsvc.go#L142
	var (
		service     = service.New(&logger)
		endpoints   = endpoint.New(service, &logger)
		httpHandler = transporthttp.NewHTTPHandler(endpoints, &logger)
		grpcServer  = transportgrpc.NewGRPCServer(endpoints, &logger)
	)

	var wg sync.WaitGroup

	{
		// The HTTP listener mounts the HTTP handler we created.
		wg.Add(1)
		go func() {
			defer wg.Done()

			srv := &http.Server{
				Addr:         fmt.Sprintf(":%d", config.HTTPServerPort),
				Handler:      httpHandler,
				ReadTimeout:  config.HTTPServerReadTimeout,
				WriteTimeout: config.HTTPServerWriteTimeout,
			}

			closeIdleConnections := make(chan struct{})
			go func() {
				sigint := make(chan os.Signal, 1)
				signal.Notify(sigint, os.Interrupt)
				<-sigint

				if err := srv.Shutdown(context.Background()); err != nil {
					log.Printf("Error to HTTP server Shutdown: %v", err)
				}
				close(closeIdleConnections)
			}()

			log.Println("HTTP server listen on", fmt.Sprintf(":%d", config.HTTPServerPort))

			err = srv.ListenAndServe()
			if !errors.Is(err, http.ErrServerClosed) {
				log.Fatalf("HTTP server ListenAndServe: %v", err)
			}

			log.Println("Initializing shutdown HTTP server...")
			<-closeIdleConnections
			log.Println("HTTP server closed with success!")
		}()
	}

	{
		// The gRPC listener mounts the gRPC server we created.
		grpcListener, err := net.Listen("tcp", "8082")
		if err != nil {
			logger.Println("transport", "gRPC", "during", "Listen", "err", err)
			os.Exit(1)
		}
		wg.Add(1)
		go func() error {
			logger.Println("transport", "gRPC", "addr", "8082")
			// we add the Go Kit gRPC Interceptor to our gRPC service as it is used by
			// the here demonstrated zipkin tracing middleware.
			baseServer := grpc.NewServer()
			pb.RegisterBurnItServer(baseServer, grpcServer)
			return baseServer.Serve(grpcListener)
		}()
	}

}
