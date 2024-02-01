# Start from the official Golang image to build the binary
FROM golang:alpine AS builder

# Install git, required for fetching Go dependencies
RUN apk add --no-cache git

# Set the working directory inside the container
WORKDIR /app

# Copy the source from the current directory to the working directory inside the container
COPY . .

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o rpc-check .

# Start a new stage from scratch for the running container
FROM alpine:latest  

# Install ca-certificates in case you make external HTTPS requests
RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy the pre-built binary file from the previous stage
COPY --from=builder /app/rpc-check .

# Command to run the executable
CMD ["./rpc-check"]
