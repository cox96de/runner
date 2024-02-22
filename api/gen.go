package api

//go:generate protoc  -I ../ -I . --go-patch_out=plugin=go,paths=source_relative:. --go-patch_out=plugin=go-grpc,paths=source_relative:. entity.proto server.proto
//go:generate mockgen -destination mock/server_mockgen.go -typed -package mock . ServerClient
//go:generate protoc-go-inject-tag -remove_tag_comment -input=*.pb.go
