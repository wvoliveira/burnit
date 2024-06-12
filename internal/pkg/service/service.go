package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"mime/multipart"

	"github.com/rs/xid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/wvoliveira/burnit/internal/pkg/config"
	e "github.com/wvoliveira/burnit/internal/pkg/error"
	"github.com/wvoliveira/burnit/internal/pkg/model"
)

type Service interface {
	Create(ctx context.Context, fileName string, file multipart.File, text []byte) (key string, err error)
}

func New(log *log.Logger, mongonClient mongo.Client) Service {
	var service Service
	{
		service = NewService(mongonClient)
	}
	return service
}

func NewService(mongonClient mongo.Client) Service {
	return service{
		mongoClient: mongonClient,
	}
}

type service struct {
	mongoClient mongo.Client
}

func (s service) Create(ctx context.Context, fileName string, file multipart.File, text []byte) (key string, err error) {
	coll := s.mongoClient.Database(config.MongoDBDatabase).Collection("keys")
	guid := xid.New()
	key = guid.String()

	fileContent := fileToBytes(file)

	res, err := coll.InsertOne(context.TODO(), bson.D{
		{"key", key},
		{"text", text},
		{"file_name", fileName},
		{"file", fileContent},
	})
	if err != nil {
		log.Println("Error: ", err.Error())
		return
	}

	log.Println("ID inserted:", res.InsertedID)

	data, err := json.MarshalIndent(res, "", "    ")
	if err != nil {
		log.Println("Error to marshal with indent: ", err.Error())
		return
	}

	fmt.Printf("%s\n", data)
	return
}

func (s service) Read(ctx context.Context, key string) (model.RequestBody, error) {
	coll := s.mongoClient.Database(config.MongoDBDatabase).Collection("keys")
	filter := bson.D{{"key", key}}

	var result requestBody
	err := coll.FindOne(context.TODO(), filter).Decode(&result)
	if errors.Is(err, mongo.ErrNoDocuments) {
		log.Println("MongoDB: record does not exist")
		return result, e.ErrMongoDBNoDocuments
	}

	if err != nil {
		log.Println("MongoDB:", err.Error())
		return result, e.ErrInternalServer
	}

	log.Println("Key: ", key)
	log.Println("Text: ", result.Text)
	log.Println("File name: ", result.FileName)
	log.Println("File: ", result.File)

	defer s.Delete(key)
	return result, nil
}

func (s service) Delete(key string) {
	coll := s.mongoClient.Database(config.MongoDBDatabase).Collection("keys")
	filter := bson.D{{"key", key}}

	_, err := coll.DeleteOne(context.TODO(), filter)
	if errors.Is(err, mongo.ErrNoDocuments) {
		log.Println("MongoDB:: record does not exist")
		return
	}

	if err != nil {
		log.Println("MongoDB:", err.Error())
		return
	}
}

func fileToBytes(file multipart.File) []byte {
	bytes := make([]byte, 1)
	var content []byte
	var size int

	for {
		read, err := file.Read(bytes)
		size += read
		content = append(content, bytes...)

		if err == io.EOF {
			break
		}
	}
	return content
}
