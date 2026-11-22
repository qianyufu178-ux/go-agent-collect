module github.com/yfyadmin/go-agent-collect

go 1.18

require (
	google.golang.org/grpc v1.56.3
	google.golang.org/protobuf v1.31.0
	gopkg.in/yaml.v3 v3.0.1
)

require (
	github.com/golang/protobuf v1.5.3 // indirect
	golang.org/x/net v0.9.0 // indirect
	golang.org/x/sys v0.7.0 // indirect
	golang.org/x/text v0.9.0 // indirect
	google.golang.org/genproto v0.0.0-20230525234025-438c736192d0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20230525234030-28d5490b6b19 // indirect
)

// 离线构建适配：go.yaml.in/yaml/v3 是 gopkg.in/yaml.v3 的官方新家（同一份代码、API 完全兼容）。
// 部分内网 / 无外网环境只缓存了前者，这里做一次路径替换，保证 `go build ./...` 开箱即用。
replace gopkg.in/yaml.v3 => go.yaml.in/yaml/v3 v3.0.5
