FROM golang:1.21.6 as build

#Change working directory
WORKDIR /app
RUN export GOPATH=/app
COPY ./code/identity_provider ./


RUN go mod download
RUN go build -o identity_provider .

ENTRYPOINT ["./identity_provider"]