FROM golang:1.22 AS builder
COPY . /src
WORKDIR /src
RUN go env -w GOPROXY=https://goproxy.cn,direct
RUN cd /src/engine/vm/runtime && CGO_ENABLED=0 go build -o /src/output/runtime .
RUN cd /src && make build_executor
FROM cox96de/qemu-static:8.0.2
COPY --from=builder /src/output/runtime /runner/runtime
COPY --from=builder /src/output/executor /runner/executor
ENTRYPOINT ["/runner/runtime"]
