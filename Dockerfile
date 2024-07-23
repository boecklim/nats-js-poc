# Use the official Golang image as the base image
FROM golang:1.22.5-alpine3.20 AS build-stage

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download all dependencies. Dependencies will be cached if the go.mod and go.sum files are not changed
RUN go mod download

# Copy the source from the current directory to the Working Directory inside the container
COPY . .

# Build the Go app
RUN go build -o /pubsub ./main.go

# Start a new stage from scratch
FROM scratch

WORKDIR /service

COPY --from=build-stage /pubsub /service/pubsub

EXPOSE 9090

CMD ["/service/pubsub"]
