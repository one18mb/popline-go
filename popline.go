package pln

import (
	"fmt"
	"strconv"
)

type ValueType int

const (
	Null   ValueType = iota
	Bool
	Int
	Float
	String
	Object
	Array
)

type Value struct {
	Type     ValueType
	key      string
	children []*Value
	boolVal  bool
	intVal   int64
	floatVal float64
	strVal   string
}

func NewObject() *Value     { return &Value{Type: Object} }
func NewArray() *Value      { return &Value{Type: Array} }
func NewNull() *Value       { return &Value{Type: Null} }
func NewBool(v bool) *Value { return &Value{Type: Bool, boolVal: v} }
func NewInt(v int64) *Value { return &Value{Type: Int, intVal: v} }
func NewFloat(v float64) *Value { return &Value{Type: Float, floatVal: v} }
func NewString(s string) *Value { return &Value{Type: String, strVal: s} }

func (v *Value) Key() string      { return v.key }
func (v *Value) Bool() bool       { return v.boolVal }
func (v *Value) Int() int64       { return v.intVal }
func (v *Value) Float() float64   { return v.floatVal }
func (v *Value) Str() string      { return v.strVal }
func (v *Value) Children() []*Value { return v.children }

func (v *Value) AddToObject(key string, val *Value) {
	if v.Type != Object { panic("not an object") }
	val.key = key
	v.children = append(v.children, val)
}

func (v *Value) AddToArray(val *Value) {
	if v.Type != Array { panic("not an array") }
	v.children = append(v.children, val)
}

// Equal compares two Value trees for structural equality.
func (v *Value) Equal(other *Value) bool {
	if v == nil && other == nil { return true }
	if v == nil || other == nil || v.Type != other.Type { return false }
	switch v.Type {
	case Null: return true
	case Bool: return v.boolVal == other.boolVal
	case Int:  return v.intVal == other.intVal
	case Float: return v.floatVal == other.floatVal
	case String: return v.strVal == other.strVal
	case Object, Array:
		if len(v.children) != len(other.children) { return false }
		for i, c := range v.children {
			oc := other.children[i]
			if c.key != oc.key { return false }
			if !c.Equal(oc) { return false }
		}
		return true
	}
	return false
}

func (v *Value) String() string {
	switch v.Type {
	case Null: return "null"
	case Bool: return fmt.Sprintf("%v", v.boolVal)
	case Int:  return fmt.Sprintf("%d", v.intVal)
	case Float: return fmt.Sprintf("%v", v.floatVal)
	case String: return strconv.Quote(v.strVal)
	case Object: return fmt.Sprintf("Object%v", v.children)
	case Array:  return fmt.Sprintf("Array%v", v.children)
	}
	return "?"
}
