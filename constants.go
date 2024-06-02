package main

import "os"

const MongoDBDatabase = "burnit"

var MongoDBURI = os.Getenv("MONGODB_URI")
