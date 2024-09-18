FROM golang:1.22 as builder
WORKDIR /src
COPY . /src
RUN go env -w GOPROXY=https://goproxy.cn,direct
RUN make build
FROM debian as agent
COPY --from=builder /src/output/agent /agent
ENTRYPOINT ["/agent"]