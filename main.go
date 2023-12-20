package main

import (
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/rs/xid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

const MONGODB_DATABASE = "burnit"

//go:embed icon.png index.html script.js
var files embed.FS
var dbClient mongo.Client

type requestBody struct {
	ID      primitive.ObjectID `bson:"_id"`
	Key     string             `json:"key"`
	Content string             `json:"content"`
}

type responseBody struct {
	Status  string `json:"status"`
	Data    any    `json:"data,omitempty"`
	Message string `json:"message"`
}

func main() {
	uri := os.Getenv("MONGODB_URI")
	if uri == "" {
		log.Fatal("You must set your 'MONGODB_URI' environment variable. See\n\t https://www.mongodb.com/docs/drivers/go/current/usage-examples/#environment-variable")
		os.Exit(2)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	dbClient = *client
	if err != nil {
		panic(err)
	}

	err = dbClient.Ping(ctx, readpref.Primary())
	if err != nil {
		panic(err)
	}

	defer func() {
		if err := dbClient.Disconnect(ctx); err != nil {
			panic(err)
		}
	}()

	mux := http.NewServeMux()
	mux.Handle("/", func() http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			log.Printf(
				`time=%s path=%s method=%s`,
				time.Now().Format(time.RFC3339), r.URL.Path, r.Method,
			)

			if (r.URL.Path == "/icon.png" || r.URL.Path == "/favicon.ico") && r.Method == "GET" {
				file, _ := files.ReadFile("icon.png")
				_, _ = w.Write(file)
				return
			}

			if (r.URL.Path == "/script.js") && r.Method == "GET" {
				w.Header().Add("content-type", "text/javascript;chartset=utf-8")
				file, _ := files.ReadFile("script.js")
				_, _ = w.Write(file)
				return
			}

			// Se o usuario acessar a chave que foi gerada
			// Ex.: http://localhost:8080/?key=cm13i2i21hc5m3qfqq8g
			// Devera coletar e remover do banco e retornar o conteudo para o usuario.
			key := r.URL.Query().Get("key")
			if r.URL.Path == "/" && r.Method == "GET" && key != "" {
				log.Println("URL query key: ", key)
				content, err := findKey(key)
				if err != nil {
					w.WriteHeader(404)
					w.Write([]byte(err.Error()))
					return
				}
				w.WriteHeader(200)
				w.Write([]byte(content))
				return
			}

			if r.URL.Path == "/" && r.Method == "GET" {
				data, err := files.ReadFile("index.html")
				if err != nil {
					log.Printf("Error: %s", err.Error())

					w.WriteHeader(500)
					w.Write([]byte(err.Error()))
					return
				}

				_, err = w.Write(data)
				if err != nil {
					log.Printf("Error: %s", err.Error())

					w.WriteHeader(500)
					w.Write([]byte(err.Error()))
					return
				}
				return
			}

			if r.URL.Path == "/" && r.Method == "POST" {
				log.Println("Body", r.Body)

				maxMem := 10
				bytes := make([]byte, maxMem)
				content := []byte{}
				var size int

				for {
					read, err := r.Body.Read(bytes)
					size += read
					content = append(content, bytes...)

					if size > 1000 {
						rb := responseBody{
							Status:  "error",
							Message: "Max size: 1000 bytes",
						}

						b, _ := json.Marshal(rb)
						w.WriteHeader(http.StatusBadRequest)
						w.Write(b)
						return
					}

					if err == io.EOF {
						break
					}
				}

				log.Printf("Size is %d\n", size)
				log.Printf("Content: %s\n", content)

				key, err := createKey(content)
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					w.Write([]byte(err.Error()))
					return
				}

				rb := responseBody{
					Status:  "ok",
					Message: key,
				}

				b, _ := json.Marshal(rb)
				w.WriteHeader(http.StatusOK)
				w.Write(b)
				return
			}

			w.WriteHeader(http.StatusNotFound)
		})
	}(),
	)

	srv := &http.Server{
		Addr:         ":8080",
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
	}

	closeIdleConnections := make(chan struct{})
	go func() {
		sigint := make(chan os.Signal, 1)
		signal.Notify(sigint, os.Interrupt)
		<-sigint

		if err := srv.Shutdown(context.Background()); err != nil {
			log.Printf("HTTP server Shutdown: %v", err)
		}
		close(closeIdleConnections)
	}()

	log.Println("HTTP server http://127.0.0.1:8080")
	if err := srv.ListenAndServe(); err != http.ErrServerClosed {
		log.Fatalf("HTTP server ListenAndServe: %v", err)
	}
	<-closeIdleConnections
}

func createKey(content []byte) (key string, err error) {
	coll := dbClient.Database(MONGODB_DATABASE).Collection("keys")
	guid := xid.New()
	key = guid.String()

	res, err := coll.InsertOne(context.TODO(), bson.D{{"key", key}, {"content", content}})
	if err != nil {
		log.Println("Error: ", err.Error())
		return
	}

	log.Println("ID inserted:", res.InsertedID)

	data, err := json.MarshalIndent(res, "", "    ")
	if err != nil {
		return
	}
	fmt.Printf("%s\n", data)
	return
}

func findKey(key string) (content string, err error) {
	coll := dbClient.Database(MONGODB_DATABASE).Collection("keys")

	var result requestBody

	filter := bson.D{{"key", key}}
	err = coll.FindOne(context.TODO(), filter).Decode(&result)
	if err == mongo.ErrNoDocuments {
		log.Println("record does not exist")
		return
	} else if err != nil {
		log.Fatal(err)
		return
	}

	content = result.Content
	log.Println("Key: ", key)
	log.Println("Content: ", content)
	return
}
