IMG_TAG ?= zcloudws/prometheus-mongodb-adapter:latest
build:
	mkdir dist || true
	go mod download
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o dist/prometheus-mongodb-adapter main.go

build-docker-image:
	docker build -t prometheus-mongodb-adapter .

push-docker-image:
	docker tag prometheus-mongodb-adapter ${IMG_TAG}
	docker push ${IMG_TAG}

run:
	go run main.go
