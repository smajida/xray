FROM golang:1.8

RUN apt-get update
RUN apt-get install -y -q cmake make
RUN apt-get install -y -q wget curl unzip
RUN apt-get install -y -q libavformat-dev libavcodec-dev libavfilter-dev libswscale-dev
RUN apt-get install -y -q libopencv-dev libopencv-core-dev checkinstall pkg-config yasm x264
RUN apt-get install --no-install-recommends -y -q curl build-essential ca-certificates git
RUN git clone https://github.com/minio/simd.git
RUN cd simd && cmake . -DCMAKE_INSTALL_PREFIX:PATH=/usr -DTOOLCHAIN="" -DTARGET=""
RUN cd simd && make -j4 install

WORKDIR /go/src/app
COPY . /go/src/app

RUN \
	go-wrapper download && \
	go-wrapper install

ENTRYPOINT ["xray"]
