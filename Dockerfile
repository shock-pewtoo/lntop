FROM golang:1.15

COPY root/build.sh /
COPY . /src

ENTRYPOINT /build.sh
