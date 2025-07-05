#!/bin/bash

# Exit on any error
set -e

echo "🔧 Installing dependencies..."
go mod download

echo "📚 Generating swagger docs..."
# Use the full path to swag since it's not in PATH
SWAG_PATH=$(go env GOPATH)/bin/swag
if [ ! -f "$SWAG_PATH" ]; then
    echo "Installing swag..."
    go install github.com/swaggo/swag/cmd/swag@latest
fi
$SWAG_PATH init --generalInfo cmd/main.go --output docs

echo "🧪 Running tests..."
go test -v -coverprofile=coverage.out ./...

echo "📊 Test coverage:"
go tool cover -func=coverage.out | grep total:

echo "✅ All tests passed!" 