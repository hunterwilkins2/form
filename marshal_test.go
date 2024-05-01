package form_test

import (
	"net/http"
	"testing"

	"github.com/hunterwilkins2/form"
)

func TestInvalidMarshalError(t *testing.T) {
	t.Parallel()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)
	err := form.Marshal(r, nil)
	if err.Error() != "form: Marshal(nil)" {
		t.Fatalf("expected %s; got %s", "form: Marshal(nil)", err.Error())
	}

	var s struct {
		Name string `form:"name"`
	}
	err = form.Marshal(r, s)
	if err.Error() != `form: Marshal(non-pointer struct { Name string "form:\"name\"" })` {
		t.Fatalf("expected %s; got %s", `form: Marshal(non-pointer struct { Name string "form:\"name\"" })`, err.Error())
	}

	var f float32
	err = form.Marshal(r, &f)
	if err.Error() != `form: Marshal(nil *float32)` {
		t.Fatalf("expected %s; got %s", `form: Marshal(nil *float32)`, err.Error())
	}
}

func TestMarshalTypeError(t *testing.T) {
	t.Parallel()
	type s struct {
		M map[string]string `form:"map"`
	}

	r, _ := http.NewRequest(http.MethodGet, "/", nil)
	err := form.Marshal(r, &s{M: map[string]string{"test": "123"}})
	if err == nil {
		t.Fatalf("expected error from Marshal")
	}
	if err.Error() != "form: cannot marshal map[test:123] (map[string]string) of Go struct field s.M into form data" {
		t.Fatalf("wrong error message. want=%s, got=%s", "form: cannot marshal map[test: 123] (map[string]string) of Go struct field s.M into form data", err.Error())
	}
}

func TestStructWithoutFormValue(t *testing.T) {
	t.Parallel()
	type s struct {
		A string `form:"a"`
		B string
	}

	testMarshalForm(t, &s{A: "a", B: "b"}, "a=a")
}

func TestLargeStructMarshal(t *testing.T) {
	type s struct {
		Name    string   `form:"name"`
		Age     int      `form:"age"`
		Pets    []string `form:"pets"`
		Balance float64  `form:"balance"`
	}

	testMarshalForm(t, &s{Name: "John", Age: 30, Pets: []string{"Rabbit", "Bird"}, Balance: 10.49}, "age=30&balance=10.490000&name=John&pets=Rabbit&pets=Bird")
}

func TestStringMarshal(t *testing.T) {
	t.Parallel()
	type s struct {
		A string `form:"a"`
	}

	testMarshalForm(t, &s{A: "strVal"}, "a=strVal")
}

func TestBoolMarshal(t *testing.T) {
	t.Parallel()
	type s struct {
		A bool `form:"a"`
	}

	testMarshalForm(t, &s{A: false}, "a=false")
}

func TestIntMarshal(t *testing.T) {
	t.Parallel()
	type s struct {
		A int `form:"a"`
	}

	testMarshalForm(t, &s{A: 88}, "a=88")
}

func TestUintMarshal(t *testing.T) {
	t.Parallel()
	type s struct {
		A uint `form:"a"`
	}

	testMarshalForm(t, &s{A: 5}, "a=5")
}

func TestFloatMarshal(t *testing.T) {
	t.Parallel()
	type s struct {
		A float32 `form:"a"`
	}

	testMarshalForm(t, &s{A: 5.349}, "a=5.349000")
}

func TestComplexMarshal(t *testing.T) {
	t.Parallel()
	type s struct {
		A complex64 `form:"a"`
	}

	testMarshalForm(t, &s{A: 9.421}, "a=%289.421000e%2B00%2B0.000000e%2B00i%29")
}

func TestSliceMarshal(t *testing.T) {
	t.Parallel()
	type s struct {
		A []int `form:"a"`
	}

	testMarshalForm(t, &s{A: []int{1, 2, 3}}, "a=1&a=2&a=3")
}

func TestArrayMarshal(t *testing.T) {
	t.Parallel()
	type s struct {
		A [2]int `form:"a"`
	}

	testMarshalForm(t, &s{A: [2]int{1, 2}}, "a=1&a=2")
}

func TestSliceMarshalTypeError(t *testing.T) {
	type s struct {
		A []map[string]string `form:"a"`
	}

	r, _ := http.NewRequest(http.MethodPost, "/", nil)
	err := form.Marshal(r, &s{A: []map[string]string{{"test": "123"}}})
	if err == nil {
		t.Fatalf("expected error from Marshal")
	}

	if err.Error() != "form: cannot marshal map[test:123] ([]map[string]string) of Go struct field s.A into form data" {
		t.Fatalf("wrong error message. want=%q, got=%q", "form: cannot marshal map[test:123] ([]map[string]string) of Go struct field s.A into form data", err.Error())
	}
}

func testMarshalForm(t *testing.T, i interface{}, expectedQuery string) {
	t.Helper()

	r, _ := http.NewRequest(http.MethodPost, "/", nil)
	err := form.Marshal(r, i)
	if err != nil {
		t.Fatalf("unexpected error from Marshal: %s", err)
	}

	if r.URL.RawQuery != expectedQuery {
		t.Fatalf("wrong query. want=%s, got=%s", expectedQuery, r.URL.RawQuery)
	}
}
