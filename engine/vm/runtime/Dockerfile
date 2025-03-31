FROM golang:1.23 AS builder
COPY . /src
WORKDIR /src
RUN go env -w GOPROXY=https://goproxy.io,direct
RUN cd /src/engine/vm/runtime && CGO_ENABLED=0 go build -o /src/output/runtime .
RUN cd /src && make build_executor build_executor_win
FROM cox96de/qemu-static:8.0.2
COPY --from=builder /src/output/runtime /runner/runtime
COPY --from=builder /src/output/executor /runner/executor
COPY --from=builder /src/output/executor.exe /runner/executor.exe
COPY --from=builder /src/engine/vm/runtime/windows_boot.ps1 /runner/windows_boot.ps1
RUN apt-get update && apt-get install -y genisoimage
RUN cd /runner && genisoimage -output executor.iso -joliet -rock executor.exe windows_boot.ps1
ENTRYPOINT ["/runner/runtime"]
