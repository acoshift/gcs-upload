#!/bin/bash
export GOOS=linux
export GOARCH=amd64
go build -o gcs-upload main.go config.go
