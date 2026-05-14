package pln

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"
	"time"
)

func assert(t *testing.T, ok bool, msg string) {
	t.Helper()
	if !ok { t.Fatal(msg) }
}

// ═══════════════ Unit Tests ═══════════════

func chk(t *testing.T, v *Value, err error) *Value {
	t.Helper()
	if err != nil { t.Fatal(err) }; return v
}

func TestBasicTypes(t *testing.T) {
	v, e := Unmarshal("{\nname: \"popline\"\n"); v = chk(t, v, e)
	assert(t, v.Children()[0].Str() == "popline", "string")

	v, e = Unmarshal("{\na: 42\n"); v = chk(t, v, e)
	assert(t, v.Children()[0].Int() == 42, "int")

	v, e = Unmarshal("{\na: 3.14\n"); v = chk(t, v, e)
	assert(t, v.Children()[0].Type == Float, "float")

	v, e = Unmarshal("{\na: true\nb: false\nc: null\n"); v = chk(t, v, e)
	assert(t, v.Children()[0].Bool(), "true")
	assert(t, !v.Children()[1].Bool(), "false")
	assert(t, v.Children()[2].Type == Null, "null")
}

func TestScalarRoot(t *testing.T) {
	v, e := Unmarshal("42"); v = chk(t, v, e)
	assert(t, v.Type == Int && v.Int() == 42, "int root")

	v, e = Unmarshal("3.14"); v = chk(t, v, e)
	assert(t, v.Type == Float, "float root")

	v, e = Unmarshal("\"hello\""); v = chk(t, v, e)
	assert(t, v.Type == String && v.Str() == "hello", "string root")

	v, e = Unmarshal("true"); v = chk(t, v, e)
	assert(t, v.Type == Bool && v.Bool() == true, "true root")

	v, e = Unmarshal("false"); v = chk(t, v, e)
	assert(t, v.Type == Bool && v.Bool() == false, "false root")

	v, e = Unmarshal("null"); v = chk(t, v, e)
	assert(t, v.Type == Null, "null root")

	v, e = Unmarshal("-42"); v = chk(t, v, e)
	assert(t, v.Type == Int && v.Int() == -42, "negative int root")
}

func TestNesting(t *testing.T) {
	v, e := Unmarshal("{\nouter: {\ninner: \"value\"\n"); v = chk(t, v, e)
	assert(t, v.Children()[0].Children()[0].Str() == "value", "nested")
}

func TestPop(t *testing.T) {
	v, e := Unmarshal("{\nouter: {\ninner: \"x\" 1\nmid: \"y\"\n"); v = chk(t, v, e)
	assert(t, len(v.Children()) == 2, "pop count")

	v, e = Unmarshal("{\na: {\nb: {\nc: \"deep\" 2\nx: \"top\"\n"); v = chk(t, v, e)
	assert(t, v.Children()[1].Str() == "top", "batch pop")
}

func TestStrings(t *testing.T) {
	v, e := Unmarshal("{\nmsg: \"He said: \"\"Hello\"\"\"\n"); v = chk(t, v, e)
	assert(t, v.Children()[0].Str() == "He said: \"Hello\"", "escape")
}

func TestErrors(t *testing.T) {
	for _, s := range []string{"{\nbad:key: 1\n", "{\n\"key\": 1\n"} {
		if _, err := Unmarshal(s); err == nil {
			t.Fatalf("expected error for: %s", s[:20])
		}
	}
}

func TestEmptyLines(t *testing.T) {
	// Empty line inside container should fail
	_, err := Unmarshal("{\n\nkey: 1\n")
	if err == nil { t.Fatal("expected error for empty line in container") }
}

// ═══════════════ Roundtrip ═══════════════

func TestRoundtrip(t *testing.T) {
	cases := []string{
		"{\na: 1\n",
		"{\na: {\nb: 1\nc: 2 1\nd: 3\n",
		"[\n1\n2\n3\n",
		"{\na: [\n1\n2 1\nb: true\n",
		"{\na: true\nb: false\nc: null\n",
	}
	for i, input := range cases {
		t.Run(fmt.Sprintf("rt-%d", i), func(t *testing.T) {
			v1, err := Unmarshal(input)
			if err != nil { t.Fatal(err) }
			s := Marshal(v1)
			v2, err := Unmarshal(s)
			if err != nil { t.Fatal(err) }
			if !v1.Equal(v2) { t.Fatal("roundtrip mismatch") }
		})
	}
}

func TestScalarRoundtrip(t *testing.T) {
	cases := []string{"42", "-42", "3.14", "\"hello\"", "true", "false", "null"}
	for i, input := range cases {
		t.Run(fmt.Sprintf("scalar-rt-%d", i), func(t *testing.T) {
			v1, err := Unmarshal(input)
			if err != nil { t.Fatal(err) }
			s := Marshal(v1)
			v2, err := Unmarshal(s)
			if err != nil { t.Fatal(err) }
			if !v1.Equal(v2) { t.Fatal("scalar roundtrip mismatch") }
		})
	}
}

// ═══════════════ Real Data Consistency ═══════════════

func TestRealDataConsistency(t *testing.T) {
	jsonBytes, err := os.ReadFile("test-package.json")
	if err != nil { t.Skip("test-package.json not found") }
	plnBytes, err := os.ReadFile("test-package.pln")
	if err != nil { t.Skip("test-package.pln not found") }

	jsonStr := string(jsonBytes)
	plnStr := string(plnBytes)

	var jsonObj interface{}
	if err := json.Unmarshal(jsonBytes, &jsonObj); err != nil {
		t.Fatal(err)
	}

	// PopLine parse
	plnVal, err := Unmarshal(plnStr)
	if err != nil { t.Fatal(err) }

	// Verify PopLine DOM matches JSON via JSON serialization
	plnJson := dumpsJson(plnVal)
	var plnObj interface{}
	json.Unmarshal([]byte(plnJson), &plnObj)

	jsonRef, _ := json.Marshal(jsonObj)
	plnRef, _ := json.Marshal(plnObj)
	assert(t, string(jsonRef) == string(plnRef), "PopLine vs JSON mismatch")

	// Roundtrip
	s := Marshal(plnVal)
	v2, err := Unmarshal(s)
	if err != nil { t.Fatal(err) }
	assert(t, plnVal.Equal(v2), "PopLine roundtrip mismatch")

	fmt.Printf("  data: JSON=%dB, PopLine=%dB (%.1f%%)\n",
		len(jsonStr), len(plnStr), float64(len(plnStr))/float64(len(jsonStr))*100)
}

func dumpsJson(v *Value) string {
	b, _ := json.Marshal(valueToInterface(v))
	return string(b)
}

func valueToInterface(v *Value) interface{} {
	if v == nil { return nil }
	switch v.Type {
	case Null:   return nil
	case Bool:   return v.Bool()
	case Int:    return v.Int()
	case Float:  return v.Float()
	case String: return v.Str()
	case Object:
		m := make(map[string]interface{})
		for _, c := range v.Children() {
			m[c.Key()] = valueToInterface(c)
		}
		return m
	case Array:
		a := make([]interface{}, len(v.Children()))
		for i, c := range v.Children() {
			a[i] = valueToInterface(c)
		}
		return a
	}
	return nil
}

// ═══════════════ Performance Benchmark ═══════════════

func TestBenchmark(t *testing.T) {
	jsonBytes, err := os.ReadFile("test-package.json")
	if err != nil { t.Skip("test-package.json not found") }
	plnBytes, err := os.ReadFile("test-package.pln")
	if err != nil { t.Skip("test-package.pln not found") }

	plnStr := string(plnBytes)

	var jsonObj interface{}
	json.Unmarshal(jsonBytes, &jsonObj)
	plnVal, _ := Unmarshal(plnStr)

	N := 5000
	fmt.Println("\n── Performance Benchmark (5000 iterations) ──")

	bench := func(label string, fn func()) time.Duration {
		fn()
		start := time.Now()
		for i := 0; i < N; i++ { fn() }
		elapsed := time.Since(start)
		fmt.Printf("  %-26s %8.0f ms  %8.0f us/op\n",
			label, float64(elapsed.Milliseconds()), float64(elapsed.Microseconds())/float64(N))
		return elapsed
	}

	jsSer := bench("json.Marshal", func() { json.Marshal(jsonObj) })
	plSer := bench("PopLine.Marshal", func() { Marshal(plnVal) })
	fmt.Printf("  %-26s %7.2fx\n", "PopLine/JSON", float64(plSer)/float64(jsSer))

	jsPar := bench("json.Unmarshal", func() {
		var v interface{}; json.Unmarshal(jsonBytes, &v)
	})
	plPar := bench("PopLine.Unmarshal", func() { Unmarshal(plnStr) })
	fmt.Printf("  %-26s %7.2fx\n", "PopLine/JSON", float64(plPar)/float64(jsPar))
}
