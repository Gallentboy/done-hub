# Stage 1: Build the web frontend
FROM node:24 as builder

WORKDIR /build

COPY web/package.json .
COPY web/yarn.lock .

RUN yarn --frozen-lockfile

COPY ./web .
COPY ./VERSION .
RUN DISABLE_ESLINT_PLUGIN='true' VITE_APP_VERSION=$(cat VERSION) yarn build

# Stage 2: Build the Go backend
FROM --platform=${BUILDPLATFORM} golang:1.24.6 AS builder2

# Install musl-tools for static CGo builds compatible with Alpine
RUN apt-get update && apt-get install -y musl-tools

# Set the working directory
WORKDIR /build

# Download dependencies
ADD go.mod go.sum .
RUN go mod download

# Copy the rest of the source code
COPY . .

# Copy the built frontend from the 'builder' stage
COPY --from=builder /build/build ./web/build

# Build the Go application for the target platform
ARG TARGETOS
ARG TARGETARCH
RUN CGO_ENABLED=1 GOOS=${TARGETOS} GOARCH=${TARGETARCH} CC=musl-gcc \
    go build -ldflags "-s -w -X 'done-hub/common/config.Version=$(cat VERSION)' -extldflags '-static'" \
    -tags netgo,osusergo \
    -o done-hub

# Stage 3: Create the final, minimal image
FROM alpine:latest

# Install necessary certificates and timezone data
RUN apk update && \
    apk upgrade && \
    apk add --no-cache ca-certificates tzdata && \
    update-ca-certificates 2>/dev/null || true

# Copy the built binary from the 'builder2' stage
COPY --from=builder2 /build/done-hub /

# Expose the application port
EXPOSE 3000

# Set the working directory for the application data
WORKDIR /data

# Set the entrypoint for the container
ENTRYPOINT ["/done-hub"]