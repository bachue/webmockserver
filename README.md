## Web Mock Server

Web Mock Server，用于 Unit Test 中作为虚拟的服务器，测试 HTTP 客户端的行为

### Usage

阅读 `src/webmockserver/example/example_test.go` 的代码

### Start Server

- 安装 Golang 1.8.1
- 安装 Glide
- 安装 Protobuf 3.2
- `go get -u github.com/golang/protobuf/{proto,protoc-gen-go}`

```bash
make
make example
./server --grpc-port 23304 --http-port 24403
```
