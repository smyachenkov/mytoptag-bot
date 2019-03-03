FROM golang:alpine AS builder

# Install git.
RUN apk update && apk add --no-cache git

WORKDIR $GOPATH/src/github.com/flslkxtc/mytoptag-bit/
COPY . .

# Fetch dependencies.

# Using go get.
RUN go get -d -v

# Build the binary.
RUN go build -o /go/bin/hello

# Run the binary.
ENTRYPOINT ["/go/bin/mytoptag-bot"]
