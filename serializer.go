package pln

import (
	"strconv"
	"strings"
)

type generator struct {
	buf           strings.Builder
	stack         []byte
	pendingPop    int
	needKey       bool
	awaitingValue bool
}

func Marshal(v *Value) string {
	g := &generator{}
	g.writeValue(v)
	return g.buf.String()
}

func (g *generator) writeValue(v *Value) {
	switch v.Type {
	case Object:
		g.startContainer('{')
		g.push('o')
		for _, c := range v.children {
			g.flushPop()
			g.buf.WriteString(c.key)
			g.buf.WriteString(": ")
			g.needKey = false
			g.awaitingValue = true
			g.writeValue(c)
		}
		g.pop()
		g.pendingPop++
		if g.top() == 'o' { g.needKey = true }
	case Array:
		g.startContainer('[')
		g.push('a')
		for _, c := range v.children {
			g.writeValue(c)
		}
		g.pop()
		g.pendingPop++
		if g.top() == 'o' { g.needKey = true }
	case Null:   g.putScalar("null")
	case Bool:   g.putScalar(strconv.FormatBool(v.boolVal))
	case Int:    g.putScalar(strconv.FormatInt(v.intVal, 10))
	case Float:  g.putScalar(strconv.FormatFloat(v.floatVal, 'g', -1, 64))
	case String: g.putString(v.strVal)
	}
}

func (g *generator) startContainer(ch byte) {
	if g.top() == 'o' && g.awaitingValue {
		g.buf.WriteByte(ch)
		g.awaitingValue = false
	} else {
		g.flushPop()
		g.buf.WriteByte(ch)
	}
	g.buf.WriteByte('\n')
}

func (g *generator) putScalar(s string) {
	if g.top() == 'o' {
		g.awaitingValue = false
		g.buf.WriteString(s)
		g.buf.WriteByte('\n')
		g.needKey = true
	} else {
		g.flushPop()
		g.buf.WriteString(s)
		g.buf.WriteByte('\n')
	}
}

func (g *generator) putString(s string) {
	if g.top() == 'o' {
		g.awaitingValue = false
		g.needKey = true
	} else {
		g.flushPop()
	}
	g.buf.WriteByte('"')
	for _, c := range s {
		g.buf.WriteRune(c)
		if c == '"' { g.buf.WriteByte('"') }
	}
	g.buf.WriteByte('"')
	g.buf.WriteByte('\n')
}

func (g *generator) flushPop() {
	if g.pendingPop > 0 {
		g.buf.WriteString(strconv.Itoa(g.pendingPop))
		g.buf.WriteByte(' ')
		g.pendingPop = 0
	}
}

func (g *generator) push(c byte) {
	g.stack = append(g.stack, c)
	g.needKey = (c == 'o')
	g.awaitingValue = false
}

func (g *generator) pop() byte {
	if len(g.stack) == 0 { return 0 }
	top := g.stack[len(g.stack)-1]
	g.stack = g.stack[:len(g.stack)-1]
	return top
}

func (g *generator) top() byte {
	if len(g.stack) == 0 { return 0 }
	return g.stack[len(g.stack)-1]
}
