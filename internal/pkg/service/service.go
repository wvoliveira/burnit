package service

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"mime/multipart"

	"github.com/rs/xid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/go-kit/log"
	"github.com/wvoliveira/burnit/internal/pkg/config"
	e "github.com/wvoliveira/burnit/internal/pkg/error"
)

type Service interface {
	Create(ctx context.Context, fileName string, file multipart.File, text []byte) (key string, err error)
}

func New(log *log.Logger, kvClient mongo.Client) Service {
	var service Service
	{
		service = NewService(kvClient)
	}
	return service
}

func NewService(kvClient mongo.Client) Service {
	return service{
		kvClient: kvClient,
	}
}

type service struct {
	kvClient mongo.Client
}

func (s service) Create(ctx context.Context, fileName string, file multipart.File, text []byte) (key string, err error) {
	coll := s.kvClient.Database(config.MongoDBDatabase).Collection("keys")
	guid := xid.New()
	key = guid.String()

	fileContent := fileToBytes(file)

	res, err := coll.InsertOne(context.TODO(), bson.D{
		{Key: "key", Value: guid.String()},
		{Key: "text", Value: text},
		{Key: "file_name", Value: fileName},
		{Key: "file", Value: fileContent},
	})
	if err != nil {
		return
	}

	_, err = json.MarshalIndent(res, "", "    ")
	if err != nil {
		return
	}

	return key, nil
}

func (s service) Read(ctx context.Context, key string) (primitive.ObjectID, string, []byte, string, multipart.File, error) {
	coll := s.kvClient.Database(config.MongoDBDatabase).Collection("keys")
	filter := bson.D{{Key: "key", Value: key}}

	var r = new(struct {
		ID       primitive.ObjectID `bson:"_id"`
		Key      string
		Text     []byte
		FileName string
		File     multipart.File
	})
	err := coll.FindOne(context.TODO(), filter).Decode(&r)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return r.ID, r.Key, r.Text, r.FileName, r.File, e.ErrMongoDBNoDocuments
	}

	if err != nil {
		return r.ID, r.Key, r.Text, r.FileName, r.File, e.ErrInternalServer
	}
	defer s.Delete(key)
	return r.ID, r.Key, r.Text, r.FileName, r.File, nil
}

func (s service) Delete(key string) {
	coll := s.kvClient.Database(config.MongoDBDatabase).Collection("keys")
	filter := bson.D{{Key: "key", Value: key}}

	_, err := coll.DeleteOne(context.TODO(), filter)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return
	}

	if err != nil {
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
