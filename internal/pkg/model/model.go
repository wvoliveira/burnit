package model

import (
	"mime/multipart"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type RequestBody struct {
	ID       primitive.ObjectID `bson:"_id"`
	Key      string             `json:"key"`
	Text     []byte             `json:"text"`
	FileName string             `json:"file_name"`
	File     multipart.File     `json:"file"`
}

type ResponseBody struct {
	Status  string `json:"status"`
	Data    any    `json:"data,omitempty"`
	Message string `json:"message"`
}

type ResponseBodyJSON struct {
	Status   string `json:"status"`
	Key      string `json:"key"`
	Text     string `json:"text"`
	FileName string `json:"file_name"`
	File     []byte `json:"file"`
}

type ResponseInfoBody struct {
	AppVersion    any    `json:"app_version"`
	GolangVersion string `json:"golang_version"`
}
