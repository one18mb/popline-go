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

## 测试

```bash
go test
```
