FROM golang:1.23 as builder
WORKDIR /src
COPY . /src
RUN go env -w GOPROXY=https://goproxy.cn,direct
RUN make build_server
FROM debian as server
COPY --from=builder /src/output/server /server
EXPOSE 8080
# Use docker-entrypoint.sh is better.
RUN mkdir -p /data/db /data/logs
ENTRYPOINT ["/server"]