FROM golang:1.21 as builder
WORKDIR /src
COPY . /src
RUN go env -w GOPROXY=https://goproxy.cn,direct
RUN make build
FROM alpine:3.14 as executor
COPY --from=builder /src/output/executor /executor