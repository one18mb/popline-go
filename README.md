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

测试数据：`package.json`（17011 B）→ `package.pln`（13074 B，**76.9%**），5000 次迭代

| 操作 | encoding/json | pln | 比 |
|------|--------------|------|------|
| 解析 | 1785 ms (357 µs/op) | 1232 ms (246 µs/op) | **0.69x** |
| 序列化 | 1522 ms (304 µs/op) | 503 ms (101 µs/op) | **0.33x** |

## 测试

```bash
go test
```
