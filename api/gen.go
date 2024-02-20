package api

//go:generate protoc  -I ../ -I . --go-patch_out=plugin=go,paths=source_relative:. --go-patch_out=plugin=go-grpc,paths=source_relative:. *.proto
//go:generate mockgen -destination mock/server_mockgen.go -package mock . ServerClient
