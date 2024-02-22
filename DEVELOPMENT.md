# Prerequest
There's some software you should install before develop this project.
- [protobuf compiler](https://grpc.io/docs/protoc-installation/)
- [protoc-gen-go & protoc-gen-go-grpc](https://grpc.io/docs/languages/go/quickstart/)
```
go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.28
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.2
```
- [gomock](https://github.com/uber-go/mock)
```
go install go.uber.org/mock/mockgen@latest
```
- [golangci-lint](https://github.com/golangci/golangci-lint)
```
go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.55.2
```
- [protopatch](https://github.com/alta/protopatch)
```
go install github.com/alta/protopatch/cmd/protoc-gen-go-patch@latest
```
- [protoc-go-inject-tag](https://github.com/favadi/protoc-go-inject-tag.git)