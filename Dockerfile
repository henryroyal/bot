FROM golang:1.10-alpine

ENV BOT_NAME chatty
ENV BOT_VERSION 0.0.1
ENV BOT_HOST www.hwr.io
ENV BOT_PORT 2022

ENV BOT_PRIVATE_KEY ./.ssh/id_rsa
ENV HOST_PUBLIC_KEY ./.ssh/id_rsa.host.pub
ENV ALLOW_INSECURE_HOSTKEY false

ENV HISTORY_PLAYBACK_LEN 20

WORKDIR /go/src/app
COPY . .

RUN go get -d -v .
RUN go install -v .

CMD ["/go/bin/app"]
