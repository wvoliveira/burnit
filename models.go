package main

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type requestBody struct {
	ID       primitive.ObjectID `bson:"_id"`
	Key      string             `json:"key"`
	Text     string             `json:"text"`
	FileName string             `json:"file_name"`
	File     []byte             `json:"file"`
}

type responseBody struct {
	Status  string `json:"status"`
	Data    any    `json:"data,omitempty"`
	Message string `json:"message"`
}

type responseBodyJSON struct {
	Status   string `json:"status"`
	Key      string `json:"key"`
	Text     string `json:"text"`
	FileName string `json:"file_name"`
	File     []byte `json:"file"`
}

type responseInfoBody struct {
	AppVersion    any    `json:"app_version"`
	GolangVersion string `json:"golang_version"`
}
