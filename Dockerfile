# Use the golang image with latest tagas the base image
FROM golang:1.22.0 as build

# Set the working directory inside the container
WORKDIR /app

# Copy the source code into the container
COPY . .

# Build the Go application
RUN CGO_ENABLED=0 GOOS=linux go build -o myapp cmd/main.go

# Start a new stage from scratch
FROM alpine:latest

# Set the working directory inside the container
WORKDIR /app

# Copy the built executable from the previous stage
COPY --from=build /app/myapp .

# Expose port on which application runs
EXPOSE 8090

# Run the application
ENTRYPOINT ["/app/myapp"]

