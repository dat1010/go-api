# Build Stage: Compile the Go application using Go 1.24.1
FROM golang:1.24.1-alpine AS build
WORKDIR /app

# Install swag CLI tool for generating Swagger documentation
RUN go install github.com/swaggo/swag/cmd/swag@latest
RUN which swag || echo "swag not found in PATH"
RUN swag version || echo "swag version failed"

# Copy dependency files and download modules
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the code
COPY . .

# Generate Swagger documentation with explicit path
RUN export PATH=$PATH:$(go env GOPATH)/bin && swag init --generalInfo cmd/main.go --output docs --parseDependency --parseInternal
RUN ls -la docs/ || echo "docs directory not found"
RUN cat docs/swagger.json | jq '.paths | keys' || echo "Failed to read swagger.json"

# Build the binary; disable CGO for a statically linked binary
RUN CGO_ENABLED=0 GOOS=linux go build -a -o main ./cmd/

# Final Stage: Run the binary in a minimal container
FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/

# Create directory for SSL certificates
RUN mkdir -p /etc/ssl/certs /etc/ssl/private

# Copy the compiled binary from the build stage
COPY --from=build /app/main .
# Copy the generated docs folder to serve Swagger documentation
COPY --from=build /app/docs ./docs

# Expose the port your Gin app listens on
EXPOSE 8080
EXPOSE 80
EXPOSE 443

# Start the application
CMD ["./main"]