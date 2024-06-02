package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/rs/xid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"io"
	"log"
	"mime/multipart"
)

func iconFile() ([]byte, error) {
	data, err := files.ReadFile("icon.png")
	if err != nil {
		log.Println("Error to read icon file:", err.Error())
		return nil, ErrInternalServer
	}
	return data, nil
}

func scriptFile() ([]byte, error) {
	data, err := files.ReadFile("script.js")
	if err != nil {
		log.Println("Error to read script file:", err.Error())
		return nil, ErrInternalServer
	}
	return data, nil
}

func indexFile() ([]byte, error) {
	data, err := files.ReadFile("index.html")
	if err != nil {
		log.Println("Error to read index file:", err.Error())
		return nil, ErrInternalServer
	}
	return data, nil
}

func keyContent(key string) (requestBody, error) {
	coll := dbClient.Database(MongoDBDatabase).Collection("keys")
	filter := bson.D{{"key", key}}

	var result requestBody
	err := coll.FindOne(context.TODO(), filter).Decode(&result)
	if errors.Is(err, mongo.ErrNoDocuments) {
		log.Println("MongoDB: record does not exist")
		return result, ErrMongoDBNoDocuments
	}

	if err != nil {
		log.Println("MongoDB:", err.Error())
		return result, ErrInternalServer
	}

	log.Println("Key: ", key)
	log.Println("Text: ", result.Text)
	log.Println("File name: ", result.FileName)
	log.Println("File: ", result.File)

	defer keyDelete(key)
	return result, nil
}

func keyDelete(key string) {
	coll := dbClient.Database(MongoDBDatabase).Collection("keys")
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

func createKey(fileName string, file multipart.File, text []byte) (key string, err error) {
	coll := dbClient.Database(MongoDBDatabase).Collection("keys")
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
