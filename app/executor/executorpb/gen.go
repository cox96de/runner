package executorpb

//go:generate protoc  -I ../../.. -I . --go-patch_out=plugin=go,paths=source_relative:. --go-patch_out=plugin=go-grpc,paths=source_relative:. service.proto
//go:generate mockgen -destination mock/mockgen.go -package mock . ExecutorClient
