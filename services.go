package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/rs/xid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"log"
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

func keyContent(key string) (string, error) {
	coll := dbClient.Database(MongoDBDatabase).Collection("keys")
	filter := bson.D{{"key", key}}

	var result requestBody
	err := coll.FindOne(context.TODO(), filter).Decode(&result)
	if errors.Is(err, mongo.ErrNoDocuments) {
		log.Println("MongoDB: record does not exist")
		return "", ErrMongoDBNoDocuments
	}

	if err != nil {
		log.Println("MongoDB:", err.Error())
		return "", ErrInternalServer
	}

	log.Println("Key: ", key)
	log.Println("Content: ", result.Content)

	defer keyDelete(key)
	return result.Content, nil
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

func createKey(content []byte) (key string, err error) {
	coll := dbClient.Database(MongoDBDatabase).Collection("keys")
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
		log.Println("Error to marshal with indent: ", err.Error())
		return
	}

	fmt.Printf("%s\n", data)
	return
}
