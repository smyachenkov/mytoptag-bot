FROM golang:alpine
RUN apk update && apk add --no-cache git
WORKDIR $GOPATH/src/github.com/flslkxtc/mytoptag-bot/
COPY . .
RUN go get -d -v
RUN go build -o /go/bin/mytoptag-bot
ENTRYPOINT ["/go/bin/mytoptag-bot"]
