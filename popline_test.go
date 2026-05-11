package popline

import (
	"testing"
)

func must(t *testing.T, v *Value, err error) *Value {
	t.Helper()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	return v
}

func TestBasicTypes(t *testing.T) {
	v := must(t, Loads("{\nname: \"popline\"\n"))
	if v.Type != Object { t.Fatal("not object") }
	if v.Children()[0].String() != "popline" { t.Fatal("string mismatch") }

	v = must(t, Loads("{\na: 42\n"))
	if v.Children()[0].Int() != 42 { t.Fatal("int mismatch") }

	v = must(t, Loads("{\na: 3.14\n"))
	if v.Children()[0].Type != Float { t.Fatal("not float") }

	v = must(t, Loads("{\na: true\nb: false\nc: null\n"))
	kids := v.Children()
	if kids[0].Bool() != true { t.Fatal("true") }
	if kids[1].Bool() != false { t.Fatal("false") }
	if kids[2].Type != Null { t.Fatal("null") }
}

func TestNesting(t *testing.T) {
	v := must(t, Loads("{\nouter: {\ninner: \"value\"\n"))
	if v.Children()[0].Children()[0].String() != "value" { t.Fatal("nested") }

	v = must(t, Loads("[\n[\n1\n2\n1 [\n3\n"))
	if len(v.Children()) != 2 { t.Fatal("arr count") }
}

func TestPop(t *testing.T) {
	v := must(t, Loads("{\nouter: {\ninner: \"x\"\n1 mid: \"y\"\n"))
	if len(v.Children()) != 2 { t.Fatal("pop count") }
	if v.Children()[1].String() != "y" { t.Fatal("pop val") }

	v = must(t, Loads("{\na: {\nb: {\nc: \"deep\"\n2 x: \"top\"\n"))
	if v.Children()[1].String() != "top" { t.Fatal("batch pop") }
}

func TestStrings(t *testing.T) {
	v := must(t, Loads("{\nmsg: \"He said: \"\"Hello\"\"\"\n"))
	if v.Children()[0].String() != "He said: \"Hello\"" { t.Fatal("escape") }

	v = must(t, Loads("{\nmsg: \"你好世界\"\n"))
	if v.Children()[0].String() != "你好世界" { t.Fatal("chinese") }
}

func TestKeys(t *testing.T) {
	v := must(t, Loads("{\nmy-key: 1\n中文键: 2\n"))
	if len(v.Children()) != 2 { t.Fatal("key count") }
}

func TestErrors(t *testing.T) {
	tests := []string{"42\n", "\"str\"\n", "true\n", "{\nbad:key: 1\n", "{\n\"key\": 1\n"}
	for _, s := range tests {
		_, err := Loads(s)
		if err == nil { t.Fatalf("expected error for: %s", s[:min(len(s), 20)]) }
	}
}

func TestRoundtrip(t *testing.T) {
	cases := []struct{
		name string
		input string
	}{
		{"simple", "{\na: 1\n"},
		{"nested", "{\na: {\nb: 1\nc: 2\n1 d: 3\n"},
		{"array", "[\n1\n2\n3\n"},
		{"mixed", "{\na: [\n1\n2\n1 b: true\n"},
		{"boolnull", "{\na: true\nb: false\nc: null\n"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			v1, err := Loads(c.input)
			if err != nil { t.Fatal(err) }
			s := Dumps(v1)
			v2, err := Loads(s)
			if err != nil { t.Fatal(err) }
			if !v1.Equal(v2) { t.Fatal("roundtrip mismatch") }
		})
	}
}

func TestComplexRoundtrip(t *testing.T) {
	input := "{\nname: \"test\"\nversion: 2\nactive: true\ntags: [\n\"web\"\n\"primary\"\n1 nested: {\nkey: \"val\"\n1 msg: \"He said: \"\"Hi\"\"\"\n"
	v1, err := Loads(input)
	if err != nil { t.Fatal(err) }
	s := Dumps(v1)
	v2, err := Loads(s)
	if err != nil { t.Fatal(err) }
	if !v1.Equal(v2) { t.Fatal("complex roundtrip mismatch") }
}
