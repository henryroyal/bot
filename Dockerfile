FROM golang:1.10-alpine

WORKDIR /go/src/app
COPY . .

RUN go get -d -v .
RUN go install -v .
RUN go build -o /usr/local/bin/bot main.go

CMD ["/usr/local/bin/bot"]
