FROM golang:latest

RUN apt-get update
RUN apt-get install -y -q wget curl unzip
RUN apt-get install -y -q libavformat-dev libavcodec-dev libavfilter-dev libswscale-dev
RUN apt-get install -y -q libopencv-dev libopencv-core-dev checkinstall pkg-config yasm x264
RUN apt-get install --no-install-recommends -y -q curl build-essential ca-certificates git mercurial bzr

WORKDIR /go/src/app

COPY . /go/src/app

RUN \
	go-wrapper download && \
	CFLAGS=-O0 go-wrapper install

ENTRYPOINT ["xray"]