FROM golang:1.23 as builder
WORKDIR /src
COPY . /src
RUN go env -w GOPROXY=https://goproxy.cn,direct
RUN make build_agent
FROM alpine:3.14 as agent
COPY --from=builder /src/output/agent /agent