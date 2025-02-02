FROM golang:1.23 as builder
WORKDIR /src
COPY . /src
RUN go env -w GOPROXY=https://goproxy.cn,direct
RUN make build_agent
FROM debian as agent
COPY --from=builder /src/output/agent /agent
ENTRYPOINT ["/agent"]