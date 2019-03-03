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

############################
# STEP 2 build a small image
############################
FROM scratch

# Copy our static executable.
COPY --from=builder /go/bin/mytoptag-bot /go/bin/mytoptag-bot

# Run the hello binary.
ENTRYPOINT ["/go/bin/mytoptag-bot"]
