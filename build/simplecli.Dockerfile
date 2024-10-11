FROM golang:1.23 as builder
WORKDIR /src
COPY . /src
RUN go env -w GOPROXY=https://goproxy.cn,direct
RUN CGO_ENABLED=0 go build -o ./output/simplecli ./cmd/simplecli
FROM alpine:3.14 as agent
COPY --from=builder /src/output/simplecli /simplecli
COPY ./cmd/simplecli/examples /examples
ENTRYPOINT ["/simplecli"]