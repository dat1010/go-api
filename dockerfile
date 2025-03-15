# Build Stage: Compile the Go application using Go 1.24.1
FROM golang:1.24.1-alpine AS build
WORKDIR /app
# Copy dependency files and download modules
COPY go.mod go.sum ./
RUN go mod download
# Copy the rest of the code and build
COPY . .
# Build the binary; disable CGO for a statically linked binary
RUN CGO_ENABLED=0 GOOS=linux go build -a -o main ./cmd/

# Final Stage: Run the binary in a minimal container
FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
# Copy the compiled binary from the build stage
COPY --from=build /app/main .
# Expose the port your Gin app listens on (e.g., 8080)
EXPOSE 8080
# Start the application
CMD ["./main"]