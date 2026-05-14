# PopLine Go

Go implementation of the PopLine serialization format.

## Install

```bash
go get github.com/one18mb/popline-go
```

## Usage

```go
import "github.com/one18mb/popline-go" // package pln

// Parse
v, err := pln.Unmarshal("{\nkey: \"value\"\n")

// Serialize
s := pln.Marshal(v)

// Build DOM
obj := pln.NewObject()
obj.AddToObject("name", pln.NewString("test"))
```

## Performance

Data: `test.json` (17011 B) → `test.pln` (13074 B, **76.9%**), 5000 iterations

| Operation | encoding/json | popline | Ratio |
|-----------|--------------|---------|-------|
| Parse | 1929 ms (385 µs/op) | 1308 ms (261 µs/op) | **0.68x** |
| Serialize | 1486 ms (297 µs/op) | 552 ms (110 µs/op) | **0.37x** |

## Test

```bash
go test
```

## Acknowledgments
This project was developed with the assistance of:
- [Claude Code](https://claude.ai) (Anthropic)
- [DeepSeek](https://deepseek.com) (DeepSeek)
