export PATH := /usr/local/go/bin:$(PATH)

SHELL := /bin/bash
.DEFAULT_GOAL := run

run:
	source .env.example && go run *.go
