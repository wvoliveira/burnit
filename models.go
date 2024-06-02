package main

import "go.mongodb.org/mongo-driver/bson/primitive"

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
