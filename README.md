# Fleek Technical Test

The purpose of this program is to create a mirror directory to encrypt data that is placed in the source directory.
The program encrypts the files using the AES CTR algorithm with a unique 256-bit key for each file, as well as a unique identifier in the form of a SHA-256 hash of the contents and relative path of the file to avoid collisions with files with similar contents.

The program run better on UNIX systems, but it can be executed inside container.

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
./fleektest -p 8080 -src ./data/unencrypted -dst ./data/encrypted
```

### Docker

Or if you prefer to run it inside docker container.

```shell script
docker build -t fleektest .
docker run --rm --name fleektest -p 80:80 fleektest
```

Since Inotify don't properly work with mounted volumes inside container, you can test by manually create files inside container.

```shell script
docker exec -it fleektest sh
cd /data/unencrypted
touch test.txt
# you can use 'vi' to set the content of file
```

## Use

If you haven't changed the default HTTP port, you can visit http://localhost/ to access the frontend of the application.
The list is refreshed every 2 seconds, you can now drag'n'drop some files inside the source directory to see their encrypted copy in the list.
The download button already embed the decryption key in the link, but if you remove it, a new page asking you to enter the key.

`/file/{HASH}?key={KEY}&filename={FILENAME}`

`filename` is optional, without it, the filename will be the hash, but it's more convenient to have it when you download the file on your disk to be recognized natively.

### Known bugs

- When large file is placed inside source directory, Inotify trigger an event before the file is completely written on disk, I have put *quick & dirty* hack for small files by sleeping 1 second before processing the file, but this doesn't work for large files. I need to find a better solution for this part. When large file fully written in source directory, you can rename it to have it correctly encrypted.

- Inotify don't seem to work with mounted volumes inside Docker container.