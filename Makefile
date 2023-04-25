.PHONY:
include .env

build:
	go build -o apiServer cmd/main.go

run: build
	./apiServer

