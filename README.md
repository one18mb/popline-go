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

测试数据：`package.json`（17011 字节） / `package.pln`（13074 字节，76.9%）

| 操作 | encoding/json | pln | 比 |
|------|--------------|------|------|
| 解析 | 1785 ms | 1232 ms | **0.69x** |
| 序列化 | 1522 ms | 503 ms | **0.33x** |

## 测试

```bash
go test
```
