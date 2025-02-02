FROM golang:1.23 as builder
WORKDIR /src
COPY . /src
RUN go env -w GOPROXY=https://goproxy.cn,direct
RUN make build_executor
FROM alpine:3.14 as executor
COPY --from=builder /src/output/executor /executor