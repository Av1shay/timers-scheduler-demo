#!/bin/bash

echo "running server tests..."
go test server/model.go server/server.go server/server_test.go -v

echo "running task service tests..."
go test task/model.go task/service.go task/service_test.go -v