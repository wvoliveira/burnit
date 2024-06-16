package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	kitgrpc "github.com/go-kit/kit/transport/grpc"
	"github.com/go-kit/log"

	"google.golang.org/grpc"

	"github.com/oklog/oklog/pkg/group"

	"github.com/wvoliveira/burnit/internal/pkg/config"
	"github.com/wvoliveira/burnit/internal/pkg/database"
	"github.com/wvoliveira/burnit/internal/pkg/endpoint"
	"github.com/wvoliveira/burnit/internal/pkg/service"
	transporthgrpc "github.com/wvoliveira/burnit/internal/pkg/transport/grpc"
	transporthttp "github.com/wvoliveira/burnit/internal/pkg/transport/http"
	"github.com/wvoliveira/burnit/pb"
)

func init() {
	if config.MongoDBURI == "" {
		fmt.Printf("You must set your 'MONGODB_URI' environment variable. " +
			"See\n\t https://www.mongodb.com/docs/drivers/go/current/usage-examples/#environment-variable")
		os.Exit(1)
	}
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), config.MongoDBTimeout)
	defer cancel()

	kvClient, err := database.NewMongoDB(ctx, config.MongoDBURI)
	if err != nil {
		fmt.Printf("Error to connect to MongoDB: %s\n", err.Error())
		os.Exit(1)
	}

	defer func() {
		if err := kvClient.Disconnect(ctx); err != nil {
			fmt.Printf("Error to disconnect from MongoDB: %s\n", err.Error())
			os.Exit(1)
		}
	}()

	var logger = log.NewLogfmtLogger(log.NewSyncWriter(os.Stderr))

	// Build the layers of the service "onion" from the inside out. First, the
	// business logic service; then, the set of endpoints that wrap the service;
	// and finally, a series of concrete transport adapters. The adapters, like
	// the HTTP handler or the gRPC server, are the bridge between Go kit and
	// the interfaces that the transports expect. Note that we're not binding
	// them to ports or anything yet; we'll do that next.
	// Copy from: https://github.com/go-kit/examples/blob/master/addsvc/cmd/addsvc/addsvc.go#L142
	var (
		service     = service.New(&logger, *kvClient)
		endpoints   = endpoint.New(service, logger)
		httpHandler = transporthttp.NewHTTPHandler(endpoints, logger)
		grpcServer  = transporthgrpc.NewGRPCServer(endpoints, logger)
	)

	// Now we're to the part of the func main where we want to start actually
	// running things, like servers bound to listeners to receive connections.
	//
	// The method is the same for each component: add a new actor to the group
	// struct, which is a combination of 2 anonymous functions: the first
	// function actually runs the component, and the second function should
	// interrupt the first function and cause it to return. It's in these
	// functions that we actually bind the Go kit server/handler structs to the
	// concrete transports and run them.
	//
	// Putting each component into its own block is mostly for aesthetics: it
	// clearly demarcates the scope in which each listener/socket may be used.
	var g group.Group

	{
		// The HTTP listener mounts the Go kit HTTP handler we created.
		httpListener, err := net.Listen("tcp", config.HTTPServerPort)
		if err != nil {
			logger.Log("transport", "HTTP", "during", "Listen", "err", err)
			os.Exit(1)
		}
		g.Add(func() error {
			logger.Log("transport", "HTTP", "addr", config.HTTPServerPort)
			return http.Serve(httpListener, httpHandler)
		}, func(error) {
			httpListener.Close()
		})
	}

	{
		// The gRPC listener mounts the Go kit gRPC server we created.
		grpcListener, err := net.Listen("tcp", config.GRPCServerPort)
		if err != nil {
			logger.Log("transport", "gRPC", "during", "Listen", "err", err)
			os.Exit(1)
		}
		g.Add(func() error {
			logger.Log("transport", "gRPC", "addr", config.GRPCServerPort)
			// we add the Go Kit gRPC Interceptor to our gRPC service as it is used by
			// the here demonstrated zipkin tracing middleware.
			baseServer := grpc.NewServer(grpc.UnaryInterceptor(kitgrpc.Interceptor))
			pb.RegisterBurnItServer(baseServer, grpcServer)
			return baseServer.Serve(grpcListener)
		}, func(error) {
			grpcListener.Close()
		})
	}

	{
		// This function just sits and waits for ctrl-C.
		cancelInterrupt := make(chan struct{})
		g.Add(func() error {
			c := make(chan os.Signal, 1)
			signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
			select {
			case sig := <-c:
				return fmt.Errorf("received signal %s", sig)
			case <-cancelInterrupt:
				return nil
			}
		}, func(error) {
			close(cancelInterrupt)
		})
	}
	logger.Log("exit", g.Run())
}
