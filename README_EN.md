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

Data: `package.json` (17011 B) / `package.pln` (13074 B, 76.9%)

| Operation | encoding/json | popline | Ratio |
|-----------|--------------|---------|-------|
| Parse | 1785 ms | 1232 ms | **0.69x** |
| Serialize | 1522 ms | 503 ms | **0.33x** |

## Test

```bash
go test
```
