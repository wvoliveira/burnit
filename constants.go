package main

import (
	"os"
	"time"
)

const AppVersion = "0.0.1"

const ServerPort = 8080
const ServerReadTimeout = 10 * time.Second
const ServerWriteTimeout = 10 * time.Second

const MongoDBTimeout = 10 * time.Second
const MongoDBDatabase = "burnit"

var MongoDBURI = os.Getenv("MONGODB_URI")
