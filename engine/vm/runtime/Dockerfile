FROM golang:1.20 AS builder
COPY . /src
WORKDIR /src
RUN go env -w GOPROXY=https://goproxy.cn,direct
RUN cd /src/engine/vm/runtime && CGO_ENABLED=0 go build -o /src/runtime .
FROM cox96de/qemu-static:8.0.2
RUN apt update && apt install -y genisoimage
COPY --from=builder /src/runtime /opt
ENTRYPOINT ["/opt/runtime"]
