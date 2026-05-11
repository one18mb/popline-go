package popline

import (
	"fmt"
	"strconv"
	"strings"
)

type parser struct {
	frames    []*Value
	key       string
	strbuf    strings.Builder
	inString  bool
}

func Loads(text string) (*Value, error) {
	p := &parser{}
	return p.parse(text)
}

func (p *parser) parse(text string) (*Value, error) {
	p.frames = nil
	p.key = ""
	p.strbuf.Reset()
	p.inString = false

	var root *Value
	lines := strings.Split(text, "\n")

	for _, line := range lines {
		line = strings.TrimSuffix(line, "\r")

		if p.inString {
			v, done := p.handleStringLine(line)
			if done {
				p.inString = false
				p.strbuf.Reset()
				if len(p.frames) == 0 { root = v; continue }
				top := p.frames[len(p.frames)-1]
				if top.Type == Object {
					top.AddToObject(p.key, v)
				} else {
					top.AddToArray(v)
				}
			}
			continue
		}

		if len(line) == 0 { continue }

		// pop prefix
		nPop := 0
		valueStart := 0
		i := 0
		for i < len(line) && line[i] >= '0' && line[i] <= '9' { i++ }
		if i > 0 && i < len(line) && line[i] == ' ' {
			nPop, _ = strconv.Atoi(line[:i])
			valueStart = i + 1
		}

		if nPop > len(p.frames) {
			return nil, fmt.Errorf("pop %d exceeds depth %d", nPop, len(p.frames))
		}
		p.frames = p.frames[:len(p.frames)-nPop]

		rest := line[valueStart:]
		if len(rest) == 0 {
			return nil, fmt.Errorf("bare pop line not allowed")
		}

		if len(p.frames) == 0 {
			var v *Value
			if rest == "{" { v = NewObject()
			} else if rest == "[" { v = NewArray()
			} else { return nil, fmt.Errorf("top level must be object or array") }
			root = v
			p.frames = append(p.frames, v)
			continue
		}

		top := p.frames[len(p.frames)-1]
		if top.Type == Object {
			err := p.parseObjectLine(rest)
			if err != nil { return nil, err }
		} else {
			err := p.parseArrayLine(rest)
			if err != nil { return nil, err }
		}
	}

	return root, nil
}

func (p *parser) parseObjectLine(rest string) error {
	sep := strings.Index(rest, ": ")
	if sep < 0 {
		return fmt.Errorf("object line must be 'key: value': %s", rest)
	}
	key := rest[:sep]
	if !isKeyValid(key) {
		return fmt.Errorf("invalid key: %s", key)
	}
	valPart := rest[sep+2:]

	top := p.frames[len(p.frames)-1]
	if valPart == "{" {
		obj := NewObject()
		top.AddToObject(key, obj)
		p.frames = append(p.frames, obj)
	} else if valPart == "[" {
		arr := NewArray()
		top.AddToObject(key, arr)
		p.frames = append(p.frames, arr)
	} else {
		p.key = key
		val, err := parseScalar(valPart, p)
		if err != nil { return err }
		if val != nil {
			top.AddToObject(key, val)
		}
	}
	return nil
}

func (p *parser) parseArrayLine(rest string) error {
	top := p.frames[len(p.frames)-1]
	if rest == "{" {
		obj := NewObject()
		top.AddToArray(obj)
		p.frames = append(p.frames, obj)
	} else if rest == "[" {
		arr := NewArray()
		top.AddToArray(arr)
		p.frames = append(p.frames, arr)
	} else {
		val, err := parseScalar(rest, p)
		if err != nil { return err }
		if val != nil {
			top.AddToArray(val)
		}
	}
	return nil
}

func (p *parser) handleStringLine(line string) (*Value, bool) {
	var result strings.Builder
	i := 0
	for i < len(line) {
		if line[i] == '"' {
			if i+1 < len(line) && line[i+1] == '"' {
				result.WriteByte('"'); i += 2
			} else {
				after := strings.TrimSpace(line[i+1:])
				if after != "" { return nil, false }
				p.strbuf.WriteString(result.String())
				return NewString(p.strbuf.String()), true
			}
		} else {
			result.WriteByte(line[i]); i++
		}
	}
	p.strbuf.WriteString(result.String())
	p.strbuf.WriteByte('\n')
	return nil, false
}

func parseScalar(s string, p *parser) (*Value, error) {
	if len(s) == 0 { return nil, fmt.Errorf("empty value") }

	if s[0] == '"' {
		return parseQuoted(s[1:], p)
	}

	switch s {
	case "true":  return NewBool(true), nil
	case "false": return NewBool(false), nil
	case "null":  return NewNull(), nil
	}

	if s[0] == '-' || (s[0] >= '0' && s[0] <= '9') {
		if strings.ContainsAny(s, ".eE") {
			f, err := strconv.ParseFloat(s, 64)
			if err == nil { return NewFloat(f), nil }
		} else {
			n, err := strconv.ParseInt(s, 10, 64)
			if err == nil { return NewInt(n), nil }
		}
	}

	return nil, fmt.Errorf("bare string must be quoted: %s", s)
}

func parseQuoted(content string, p *parser) (*Value, error) {
	var result strings.Builder
	i := 0
	for i < len(content) {
		if content[i] == '"' {
			if i+1 < len(content) && content[i+1] == '"' {
				result.WriteByte('"'); i += 2
			} else {
				after := strings.TrimSpace(content[i+1:])
				if after != "" { return nil, fmt.Errorf("trailing content after quote") }
				return NewString(result.String()), nil
			}
		} else {
			result.WriteByte(content[i]); i++
		}
	}
	// multi-line
	p.inString = true
	p.strbuf.Reset()
	p.strbuf.WriteString(content)
	p.strbuf.WriteByte('\n')
	return nil, nil
}

func isKeyValid(key string) bool {
	if len(key) == 0 { return false }
	for _, c := range key {
		if c == ':' || c == '"' || c == '{' || c == '}' ||
			c == '[' || c == ']' || c == '#' ||
			c == ' ' || c == '\t' || c == '\n' || c == '\r' {
			return false
		}
	}
	return true
}
