FROM golang:1.21-alpine as build-env

RUN apk add --update ca-certificates

WORKDIR /workspace

COPY . .

RUN go mod download

RUN mkdir dist &&\
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o dist/prometheus-mongodb-adapter main.go

FROM scratch
COPY --from=build-env /workspace/dist/prometheus-mongodb-adapter /bin/prometheus-mongodb-adapter
COPY --from=build-env /etc/ssl/certs /etc/ssl/certs
ENTRYPOINT [ "/bin/prometheus-mongodb-adapter" ]
