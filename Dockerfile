# Build Stage
FROM golang:1.23.0-alpine3.20 as builder

# Set the working directory for the builder stage
WORKDIR /builder

# Define a build argument
ARG GOPROXY
ARG GONOSUMDB

# Copy the source code
COPY . .

# Download Go modules
RUN go mod tidy

# Build the application
RUN CGO_ENABLED=0 go build -o proxy -ldflags="-s -w" -trimpath -o guard ./server/.

# Final Stage
FROM alpine:3.20.2

# Install openssh, because we need ssh-keygen
RUN apk add --no-cache openssh

# Set the working directory for the final stage
WORKDIR /app

# Copy the binary from the builder stage to the final stage
COPY --from=builder /builder/guard guard

# Set the entrypoint
ENTRYPOINT ["/app/guard"]