FROM golang
ADD . /go/src/github.com/flslkxtc/mytoptag-bot
RUN go get ./...
RUN go install github.com/flslkxtc/mytoptag-bot
ENTRYPOINT /go/bin/mytoptag-bot
EXPOSE 8080
