package config

import (
	"os"
	"time"
)

const AppVersion = "0.0.1"

const HTTPServerPort = ":8080"
const HTTPServerReadTimeout = 10 * time.Second
const HTTPServerWriteTimeout = 10 * time.Second

const GRPCServerPort = ":8081"

const MongoDBTimeout = 10 * time.Second
const MongoDBDatabase = "burnit"

var MongoDBURI = os.Getenv("MONGODB_URI")
