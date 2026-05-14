# PopLine Go

PopLine 序列化格式的 Go 实现。

## 安装

```bash
go get github.com/one18mb/popline-go
```

## 使用

```go
import "github.com/one18mb/popline-go" // package pln

// 解析
v, err := pln.Unmarshal("{\nkey: \"value\"\n")

// 序列化
s := pln.Marshal(v)

// 构建 DOM
obj := pln.NewObject()
obj.AddToObject("name", pln.NewString("test"))
```

## 性能

测试数据：`test-package.json`（17011 B）→ `test-package.pln`（13074 B，**76.9%**），5000 次迭代

| 操作 | encoding/json | pln | 比 |
|------|--------------|------|------|
| 解析 | 1929 ms (385 µs/op) | 1308 ms (261 µs/op) | **0.68x** |
| 序列化 | 1486 ms (297 µs/op) | 552 ms (110 µs/op) | **0.37x** |

## 测试

```bash
go test
```

## 致谢
本项目的开发得到了以下 AI 工具的大力协助：
- [Claude Code](https://claude.ai)（Anthropic）
- [DeepSeek](https://deepseek.com)（深度求索）
