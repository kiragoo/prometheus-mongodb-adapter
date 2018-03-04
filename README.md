# prometheus-mongodb-adapter: Prometheus remote storage adapter implementation for MongoDB

Now under development

## Features & TODO

- [x] Prometheus remote storage adapter implementation for MongoDB
- [ ] Prometheus 1.8 support
- [x] Prometheus 2.1 support
- [ ] MongoDB 3.0 support
- [ ] MongoDB 3.2 support
- [ ] MongoDB 3.4 support
- [x] MongoDB 3.6 support
- [ ] MongoDB 3.7 support
- [ ] Secure MongoDB support (tls connection)
- [ ] Unit test

## Getting Started

### Docker

```bash
docker run -it \
    --name prometheus-mongodb-adapter \
    -p 8080:8080 \
    sasuraiossan/prometheus-mongodb-adapter
```

### go get

```bash
# TODO
```

## Configuration

```bash
$ prometheus-mongodb-adapter --help                                                  
NAME:
   prometheus-mongodb-adapter

OPTIONS:
   --mongo-url value, -m value   (default: "mongodb://localhost:27017/prometheus") [$MONGO_URL]
   --database value, -d value    (default: "prometheus") [$DATABASE_NAME]
   --collection value, -c value  (default: "prometheus") [$COLLECTION_NAME]
   --address value, -a value     (default: "0.0.0.0:8080") [$LISTEN_ADDRESS]
   --help, -h                    show help
   --version, -v                 print the version
```

If database name is included in `mongo-url`, `databse` argument is ignored.

## License

MIT License
