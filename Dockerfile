FROM golang:alpine as build-env
RUN apk add --update git

# RUN go get github.com/gogo/protobuf/proto
# RUN go get github.com/golang/snappy
# RUN go get github.com/prometheus/prometheus/prompb
# RUN go get github.com/sirupsen/logrus
# RUN go get github.com/golang/protobuf/jsonpb
# RUN go get github.com/globalsign/mgo
# RUN go get gopkg.in/urfave/cli.v1
# RUN go get github.com/gorilla/handlers
# RUN go get github.com/julienschmidt/httprouter

COPY . /go/src/github.com/sasurai-ossan/prometheus-mongodb-adapter/
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64  go build \
    -o /go/bin/prometheus-mongodb-adapter \
    /go/src/github.com/sasurai-ossan/prometheus-mongodb-adapter/main.go

RUN apk add --update ca-certificates

FROM scratch
COPY --from=build-env /go/bin/prometheus-mongodb-adapter /bin/prometheus-mongodb-adapter
COPY --from=build-env /etc/ssl/certs /etc/ssl/certs
ENTRYPOINT [ "/bin/prometheus-mongodb-adapter" ]
