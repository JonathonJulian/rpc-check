# Start from the official Golang image to build the binary
FROM golang:alpine AS builder

# Install git, required for fetching Go dependencies
RUN apk add --no-cache git

# Set the working directory inside the container
WORKDIR /app

# Initialize a new module (only if you don't have a go.mod file)
# RUN go mod init your/module/name

# Copy the go.mod and go.sum file (if present) to fetch dependencies
COPY go.mod go.sum ./

# Download all dependencies. Dependencies will be cached if the go.mod and go.sum files are not changed
RUN go mod download

# Copy the source from the current directory to the working directory inside the container
COPY . .

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o rpc-check .

# Start a new stage from scratch for the running container
FROM alpine:latest  
LABEL org.opencontainers.image.source=https://github.com/JonathonJulian/rpc-check/
# Install ca-certificates in case you make external HTTPS requests
RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy the pre-built binary file from the previous stage
COPY --from=builder /app/rpc-check .

# Command to run the executable
CMD ["./rpc-check"]
