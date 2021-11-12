FROM golang:1.16

RUN apt update && apt upgrade -y

WORKDIR $GOPATH/src/app
COPY . .

RUN go build -ldflags="-s -w" main.go

ENV TOKEN ORG

ENTRYPOINT ["./entrypoint.sh"]
