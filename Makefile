.PHONY: all example

OUTPUT=server

all: $(OUTPUT)
example: server
	GOPATH=$(PWD) go test -parallel 1 -v webmockserver/example

server: $(shell find src -type f \( -name "*.go" ! -path "src/webmockserver/example/*" \)) src/webmockserver/proto/assertion.pb.go src/webmockserver/glide.lock
	GOPATH=$(PWD) go build -ldflags='-s -w' -o server webmockserver

src/webmockserver/glide.lock: src/webmockserver/glide.yaml
	cd src/webmockserver && GOPATH=$(PWD) glide update -v

src/webmockserver/proto/assertion.pb.go: src/webmockserver/proto/assertion.proto
	protoc --go_out=plugins=grpc:. src/webmockserver/proto/assertion.proto

clean:
	$(RM) -f $(OUTPUT) src/proto/assertion.pb.go
