package pln

import (
	"fmt"
	"strings"
)

type generator struct {
	buf           strings.Builder
	stack         []byte
	needKey       bool
	awaitingValue bool
}

func Marshal(v *Value) string {
	g := &generator{}
	g.writeValue(v, 0)
	return g.buf.String()
}

func (g *generator) writeValue(v *Value, closePop int) {
	switch v.Type {
	case Object:
		g.startContainer('{')
		g.push('o')
		n := len(v.children)
		for i, c := range v.children {
			childPop := 0
			if i == n-1 {
				childPop = closePop + 1
			}
			g.buf.WriteString(c.key)
			g.buf.WriteString(": ")
			g.needKey = false
			g.awaitingValue = true
			g.writeValue(c, childPop)
		}
		g.pop()
		if g.top() == 'o' { g.needKey = true }
	case Array:
		g.writeArrayInline(v, true, closePop)
	case Null:   g.putScalar("null", closePop)
	case Bool:   g.putScalar(fmt.Sprintf("%t", v.boolVal), closePop)
	case Int:    g.putScalar(fmt.Sprintf("%d", v.intVal), closePop)
	case Float:  g.putScalar(fmt.Sprintf("%g", v.floatVal), closePop)
	case String: g.putString(v.strVal, closePop)
	}
}

func (g *generator) writeArrayInline(v *Value, first bool, closePop int) {
	ch := byte('[')
	typ := byte('a')
	if v.Type == Object { ch = '{'; typ = 'o' }

	if first && g.top() == 'o' && g.awaitingValue {
		g.buf.WriteByte(ch)
		g.awaitingValue = false
	} else if first {
		g.buf.WriteByte(ch)
	} else {
		g.buf.WriteByte(ch)
	}

	c := v.children

	// Always use non-inline path for correct closePop propagation
	g.buf.WriteByte('\n')
	g.push(typ)
	g.needKey = (typ == 'o')
	g.awaitingValue = false
	if v.Type == Object {
		n := len(c)
		for i, c2 := range c {
			childPop := 0
			if i == n-1 {
				childPop = closePop + 1
			}
			g.buf.WriteString(c2.key)
			g.buf.WriteString(": ")
			g.needKey = false
			g.awaitingValue = true
			g.writeValue(c2, childPop)
		}
	} else {
		n := len(c)
		for i, c2 := range c {
			childPop := 0
			if i == n-1 {
				childPop = closePop + 1
			}
			g.writeValue(c2, childPop)
		}
	}
	g.pop()
	if g.top() == 'o' { g.needKey = true }
}

func (g *generator) startContainer(ch byte) {
	if g.top() == 'o' && g.awaitingValue {
		g.buf.WriteByte(ch)
		g.awaitingValue = false
	} else {
		g.buf.WriteByte(ch)
	}
	g.buf.WriteByte('\n')
}

func (g *generator) putScalar(s string, closePop int) {
	if g.top() == 'o' {
		g.awaitingValue = false
	}
	g.buf.WriteString(s)
	if closePop > 0 {
		g.buf.WriteString(fmt.Sprintf(" %d", closePop))
	}
	g.buf.WriteByte('\n')
	if g.top() == 'o' {
		g.needKey = true
	}
}

func (g *generator) putString(s string, closePop int) {
	if g.top() == 'o' {
		g.awaitingValue = false
		g.needKey = true
	}
	g.buf.WriteByte('"')
	for _, c := range s {
		g.buf.WriteRune(c)
		if c == '"' { g.buf.WriteByte('"') }
	}
	g.buf.WriteByte('"')
	if closePop > 0 {
		g.buf.WriteString(fmt.Sprintf(" %d", closePop))
	}
	g.buf.WriteByte('\n')
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
