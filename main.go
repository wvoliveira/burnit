package main

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"log"
	"net/http"
	"os"
	"os/signal"
)

//go:embed icon.png index.html script.js
var files embed.FS
var dbClient mongo.Client

func init() {
	if MongoDBURI == "" {
		log.Fatal("You must set your 'MONGODB_URI' environment variable. " +
			"See\n\t https://www.mongodb.com/docs/drivers/go/current/usage-examples/#environment-variable")
	}
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), MongoDBTimeout)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(MongoDBURI))
	if err != nil {
		log.Fatal(err)
	}

	dbClient = *client
	err = dbClient.Ping(ctx, readpref.Primary())
	if err != nil {
		log.Fatal(err)
	}

	defer func() {
		if err := dbClient.Disconnect(ctx); err != nil {
			log.Fatal(err)
		}
	}()

	router := Router()
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", ServerPort),
		Handler:      router,
		ReadTimeout:  ServerReadTimeout,
		WriteTimeout: ServerWriteTimeout,
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

	log.Println("HTTP server http://127.0.0.1:8080")

	err = srv.ListenAndServe()
	if !errors.Is(err, http.ErrServerClosed) {
		log.Fatalf("HTTP server ListenAndServe: %v", err)
	}

	log.Println("Initializing shutdown HTTP server...")
	<-closeIdleConnections
	log.Println("HTTP server closed with success!")
}
