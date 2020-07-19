# Fleek Technical Test

## Build & Run

Before running the program, you will need to create two directories, one for unencrypted content where you can put your own files and another one for encrypted mirror.

```shell script
mkdir -p data/unencrypted
mkdir -p data/encrypted
```

### Locally

You can build and run the program directly on you computer if you have Golang installed.

```shell script
go get ./...
go test ./...
go build -o fleektest ./cmd/server
./fleektest -p 8080 -src ./data/uncencrypted -dst ./data/encrypted
```

### Docker

Or if you prefer to run it inside docker container, don't forget to expose http server port and mount unencrypted/encrypted volumes to have something persistent.

```shell script
docker build -t fleektest .
docker run --rm --name fleektest -p 80:80 -v $PWD/data/unencrypted:/data/unencrypted -v $PWD/data/encrypted:/data/encrypted fleektest
```