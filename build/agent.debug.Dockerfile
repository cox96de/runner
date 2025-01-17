FROM golang:1.23 as builder
WORKDIR /src
COPY . /src
RUN go env -w GOPROXY=https://goproxy.cn,direct
RUN make build_agent_debug
RUN CGO_ENABLED=0 go install github.com/go-delve/delve/cmd/dlv@latest
FROM golang:1.23 as agent
COPY --from=builder /src/output/agent /agent
COPY --from=builder /go/bin/dlv /dlv