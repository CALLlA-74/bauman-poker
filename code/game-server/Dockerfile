FROM golang:1.21.6 as build

#Change working directory
WORKDIR /app
RUN export GOPATH=/app
COPY ./code/game-server ./


RUN go mod download
RUN go build -o game-server .

ENTRYPOINT ["./game-server"]