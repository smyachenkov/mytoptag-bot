FROM golang:latest
WORKDIR $GOPATH/src/github.com/flslkxtc/mytoptag-bot
COPY . .
RUN go get -d -v ./...
RUN go install -v ./...
EXPOSE 8080
CMD ["mytoptag-bot"]
