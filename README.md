# Xray

Deep learning based object detection for video.

## Install

```sh
go get -d github.com/minio/xray
cd $GOPATH/src/github.com/minio/xray
docker build . -t xray
```

## Run

```sh
docker run --it --rm -p 8080:8080 xray
```
