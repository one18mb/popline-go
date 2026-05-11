# PopLine Go

PopLine 序列化格式的 Go 实现。

## 安装

```bash
go get github.com/one18mb/popline-go
```

## 使用

```go
import "github.com/one18mb/popline-go"

// 解析
v, err := popline.Loads("{\nkey: \"value\"\n")

// 序列化
s := popline.Dumps(v)

// 构建 DOM
obj := popline.NewObject()
obj.AddToObject("name", popline.NewString("test"))
```

## 性能

测试数据：`package.json`（17011 字节） / `package.pln`（13074 字节，76.9%）

| 操作 | encoding/json | popline | 比 |
|------|--------------|---------|------|
| 解析 | 1785 ms | 1232 ms | **0.69x** |
| 序列化 | 1522 ms | 503 ms | **0.33x** |

## 测试

```bash
go test
```
