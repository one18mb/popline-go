package pln

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

func Unmarshal(text string) (*Value, error) {
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
			v, nPop, done := p.handleStringLine(line)
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
				if nPop > 0 {
					if nPop >= len(p.frames) {
						return nil, fmt.Errorf("suffix pop %d exceeds depth %d (would pop root)", nPop, len(p.frames))
					}
					p.frames = p.frames[:len(p.frames)-nPop]
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
			// Check top-level inline containers: `[ [` or `[ {`
			if len(rest) > 1 && rest[0] == '[' {
				rp := strings.TrimLeft(rest[1:], " \t")
				if len(rp) > 0 && (rp[0] == '[' || rp[0] == '{') {
					if err := p.inlineContainers(rest); err != nil { return nil, err }
					root = p.frames[0]
					continue
				}
			}
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

	// Check value inline containers: `key: [ [` or `key: [ {`
	if len(valPart) > 1 && (valPart[0] == '[' || valPart[0] == '{') {
		rp := strings.TrimLeft(valPart[1:], " \t")
		if len(rp) > 0 && (rp[0] == '[' || rp[0] == '{') {
			p.key = key
			return p.inlineContainers(valPart)
		}
	}

	if valPart == "{" {
		obj := NewObject()
		top.AddToObject(key, obj)
		p.frames = append(p.frames, obj)
	} else if valPart == "[" {
		arr := NewArray()
		top.AddToObject(key, arr)
		p.frames = append(p.frames, arr)
	} else {
		// Extract trailing pop suffix for leaf values
		nPop := 0
		vlen := len(valPart)
		if vlen > 0 && valPart[0] != '{' && valPart[0] != '[' {
			nPop = trimPopSuffix(valPart, &vlen)
		}
		valPart = valPart[:vlen]
		if len(valPart) == 0 {
			return nil
		}

		p.key = key
		val, err := parseScalar(valPart, p)
		if err != nil { return err }
		if val != nil {
			top.AddToObject(key, val)
		}
		if nPop > 0 {
			if nPop >= len(p.frames) {
				return fmt.Errorf("suffix pop %d exceeds depth %d (would pop root)", nPop, len(p.frames))
			}
			p.frames = p.frames[:len(p.frames)-nPop]
		}
	}
	return nil
}

func (p *parser) parseArrayLine(rest string) error {
	top := p.frames[len(p.frames)-1]

	// Check array element inline containers: `[ [`、`[ {`、`{ [`、`{ {`
	if len(rest) > 1 && (rest[0] == '[' || rest[0] == '{') {
		rp := strings.TrimLeft(rest[1:], " \t")
		if len(rp) > 0 && (rp[0] == '[' || rp[0] == '{') {
			return p.inlineContainers(rest)
		}
	}

	if rest == "{" {
		obj := NewObject()
		top.AddToArray(obj)
		p.frames = append(p.frames, obj)
	} else if rest == "[" {
		arr := NewArray()
		top.AddToArray(arr)
		p.frames = append(p.frames, arr)
	} else {
		// Extract trailing pop suffix for leaf values
		nPop := 0
		restLen := len(rest)
		if restLen > 0 && rest[0] != '{' && rest[0] != '[' {
			nPop = trimPopSuffix(rest, &restLen)
		}
		trimmedRest := rest[:restLen]
		if len(trimmedRest) == 0 {
			return nil
		}

		val, err := parseScalar(trimmedRest, p)
		if err != nil { return err }
		if val != nil {
			top.AddToArray(val)
		}
		if nPop > 0 {
			if nPop >= len(p.frames) {
				return fmt.Errorf("suffix pop %d exceeds depth %d (would pop root)", nPop, len(p.frames))
			}
			p.frames = p.frames[:len(p.frames)-nPop]
		}
	}
	return nil
}

// inlineContainers parses consecutive container openers on a single line: `[ [`, `[ {`, etc.
func (p *parser) inlineContainers(s string) error {
	part := strings.TrimSpace(s)
	for len(part) > 0 {
		ch := part[0]
		if ch != '{' && ch != '[' {
			return fmt.Errorf("inline containers must be '{' or '['")
		}
		var v *Value
		if ch == '{' {
			v = NewObject()
		} else {
			v = NewArray()
		}
		if len(p.frames) == 0 {
			p.frames = append(p.frames, v)
		} else {
			top := p.frames[len(p.frames)-1]
			if top.Type == Object && p.key != "" {
				top.AddToObject(p.key, v)
				p.key = ""
			} else if top.Type == Object {
				top.AddToObject("", v)
			} else {
				top.AddToArray(v)
			}
			p.frames = append(p.frames, v)
		}
		part = strings.TrimLeft(part[1:], " \t")
	}
	return nil
}

func (p *parser) handleStringLine(line string) (*Value, int, bool) {
	var result strings.Builder
	i := 0
	for i < len(line) {
		if line[i] == '"' {
			if i+1 < len(line) && line[i+1] == '"' {
				result.WriteByte('"'); i += 2
			} else {
				after := line[i+1:]
				nPop := popSuffixAfter(after)
				if nPop < 0 { return nil, 0, false }
				p.strbuf.WriteString(result.String())
				return NewString(p.strbuf.String()), nPop, true
			}
		} else {
			result.WriteByte(line[i]); i++
		}
	}
	p.strbuf.WriteString(result.String())
	p.strbuf.WriteByte('\n')
	return nil, 0, false
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

// trimPopSuffix extracts trailing " N" from content, returns pop count, adjusts *contentLen.
// If no valid suffix is found, *contentLen is unchanged and 0 is returned.
func trimPopSuffix(s string, contentLen *int) int {
	clen := *contentLen
	if clen < 2 {
		return 0
	}
	i := clen - 1
	if s[i] < '0' || s[i] > '9' {
		return 0
	}
	for i > 0 && s[i-1] >= '0' && s[i-1] <= '9' {
		i--
	}
	if i == 0 || s[i-1] != ' ' {
		return 0
	}
	n := 0
	for j := i; j < clen; j++ {
		n = n*10 + int(s[j]-'0')
	}
	*contentLen = i - 1
	return n
}

// popSuffixAfter validates content after closing quote: ""=0, " N"=N, other=-1.
func popSuffixAfter(s string) int {
	if len(s) == 0 {
		return 0
	}
	if s[0] != ' ' {
		return -1
	}
	if len(s) < 2 || s[1] < '0' || s[1] > '9' {
		return -1
	}
	n := 0
	for i := 1; i < len(s); i++ {
		if s[i] < '0' || s[i] > '9' {
			return -1
		}
		n = n*10 + int(s[i]-'0')
	}
	return n
}

func isKeyValid(key string) bool {
	if len(key) == 0 { return false }
	for _, c := range key {
		if c == ':' || c == '"' || c == '{' ||
			c == '[' || c == '#' ||
			c == ' ' || c == '\t' || c == '\n' || c == '\r' {
			return false
		}
	}
	return true
}
