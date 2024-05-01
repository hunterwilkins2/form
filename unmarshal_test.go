package form_test

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sort"
	"strings"
	"testing"

	"github.com/hunterwilkins2/form"
	"golang.org/x/exp/constraints"
)

func TestInvalidUnmarshalError(t *testing.T) {
	r, _ := http.NewRequest(http.MethodGet, "/", nil)
	err := form.Unmarshal(r, nil)
	if err.Error() != "form: Unmarshal(nil)" {
		t.Fatalf("expected %s; got %s", "form: Unmarshal(nil)", err.Error())
	}

	var s struct {
		Name string `form:"name"`
	}
	err = form.Unmarshal(r, s)
	if err.Error() != `form: Unmarshal(non-pointer struct { Name string "form:\"name\"" })` {
		t.Fatalf("expected %s; got %s", `form:Unmarshal(non-pointer struct { Name string "form:\"name\"" })`, err.Error())
	}

	var f float32
	err = form.Unmarshal(r, &f)
	if err.Error() != `form: Unmarshal(nil *float32)` {
		t.Fatalf("expected %s; got %s", `form: Unmarshal(nil *float32)`, err.Error())
	}
}

func TestUnwrapUnmarshalTypeError(t *testing.T) {
	t.Parallel()
	err := fmt.Errorf("test error message")

	wrappedErr := &form.UnmarshalTypeError{Err: err}
	if errors.Unwrap(wrappedErr) != err {
		t.Fatalf("wrong wrapped error. want=%s, got=%s", err, errors.Unwrap(wrappedErr))
	}
}

type UrlFormData[T constraints.Ordered] struct {
	Single T    `form:"single"`
	Slice  []T  `form:"slice"`
	Array  [2]T `form:"array"`
}

func TestUnmarshalString(t *testing.T) {
	t.Parallel()
	data := UrlFormData[string]{
		Single: "a",
		Slice:  []string{"a", "b"},
		Array:  [2]string{"a", "b"},
	}

	testUnmarshalFormData(t, data)
}

func TestUnmarshalBool(t *testing.T) {
	t.Parallel()
	type decodeBool struct {
		B bool `form:"b"`
	}

	var actual decodeBool
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		err := form.Unmarshal(r, &actual)
		if err != nil {
			t.Fatalf("unexpected unmarshal error: %s", err)
		}
	}))
	defer server.Close()

	r, err := http.NewRequest(http.MethodPost, server.URL, nil)
	if err != nil {
		t.Fatalf("unexpected error creating request: %s", err)
	}
	form := make(url.Values)
	form.Add("b", "true")
	r.URL.RawQuery = form.Encode()
	r.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(r)
	if err != nil {
		t.Fatalf("unexpected error sending request: %s", err)
	}
	defer resp.Body.Close()

	if !actual.B {
		t.Fatalf("mismatch bool value. want=%t, got=%t", true, actual.B)
	}
}

func TestUnmarshalBoolError(t *testing.T) {
	t.Parallel()
	type s struct {
		Val bool `form:"value"`
	}

	testUnmarshalFormError(t, "notABool", &s{}, "form: cannot unmarshal notABool into Go struct field s.Val of type bool: strconv.ParseBool: parsing \"notABool\": invalid syntax")
}

func TestUnmarshalInt(t *testing.T) {
	t.Parallel()
	data := UrlFormData[int]{
		Single: 1,
		Slice:  []int{5, 10},
		Array:  [2]int{1, 2},
	}

	testUnmarshalFormData(t, data)
}

func TestUnmarshalIntError(t *testing.T) {
	t.Parallel()
	type s struct {
		Val int8 `form:"value"`
	}

	testUnmarshalFormError(t, "notAInt", &s{}, "form: cannot unmarshal notAInt into Go struct field s.Val of type int8: strconv.ParseInt: parsing \"notAInt\": invalid syntax")
	testUnmarshalFormError(t, "257", &s{}, "form: cannot unmarshal 257 into Go struct field s.Val of type int8: 257 overflows int8 value")
}

func TestUnmarshalUInt(t *testing.T) {
	t.Parallel()
	data := UrlFormData[uint]{
		Single: 1,
		Slice:  []uint{5, 10},
		Array:  [2]uint{1, 2},
	}

	testUnmarshalFormData(t, data)
}

func TestUnmarshalUIntError(t *testing.T) {
	t.Parallel()
	type s struct {
		Val uint8 `form:"value"`
	}

	testUnmarshalFormError(t, "-1", &s{}, "form: cannot unmarshal -1 into Go struct field s.Val of type uint8: strconv.ParseUint: parsing \"-1\": invalid syntax")
	testUnmarshalFormError(t, "257", &s{}, "form: cannot unmarshal 257 into Go struct field s.Val of type uint8: 257 overflows uint8 value")
}

func TestUnmarshalFloat(t *testing.T) {
	t.Parallel()
	data := UrlFormData[float32]{
		Single: 1,
		Slice:  []float32{5, 10},
		Array:  [2]float32{1, 2},
	}

	testUnmarshalFormData(t, data)
}

func TestUnmarshalFloatError(t *testing.T) {
	t.Parallel()
	type s struct {
		Val float32 `form:"value"`
	}

	testUnmarshalFormError(t, "notAFloat", &s{}, "form: cannot unmarshal notAFloat into Go struct field s.Val of type float32: strconv.ParseFloat: parsing \"notAFloat\": invalid syntax")
	testUnmarshalFormError(t, "3402823000000000000000000000000000000000001", &s{}, "form: cannot unmarshal 3402823000000000000000000000000000000000001 into Go struct field s.Val of type float32: 3402823000000000000000000000000000000000001 overflows float32 value")
}

func TestUnmarshalComplex(t *testing.T) {
	t.Parallel()
	type decodeComplex struct {
		C complex64 `form:"complex"`
	}

	var actual decodeComplex
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		err := form.Unmarshal(r, &actual)
		if err != nil {
			t.Fatalf("unexpected unmarshal error: %s", err)
		}
	}))
	defer server.Close()

	r, err := http.NewRequest(http.MethodPost, server.URL, nil)
	if err != nil {
		t.Fatalf("unexpected error creating request: %s", err)
	}
	form := make(url.Values)
	form.Add("complex", "8")
	r.URL.RawQuery = form.Encode()
	r.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(r)
	if err != nil {
		t.Fatalf("unexpected error sending request: %s", err)
	}
	defer resp.Body.Close()

	if actual.C != 8 {
		t.Fatalf("mismatch bool value. want=%f, got=%f", complex64(8), actual.C)
	}
}

func TestUnmarshalComplexError(t *testing.T) {
	t.Parallel()
	type s struct {
		Val complex64 `form:"value"`
	}

	testUnmarshalFormError(t, "notAComplex", &s{}, "form: cannot unmarshal notAComplex into Go struct field s.Val of type complex64: strconv.ParseComplex: parsing \"notAComplex\": invalid syntax")
	testUnmarshalFormError(t, "1.7976931348623157e+308", &s{}, "form: cannot unmarshal 1.7976931348623157e+308 into Go struct field s.Val of type complex64: 1.7976931348623157e+308 overflows complex64 value")
}

func TestUnmarshalNoValues(t *testing.T) {
	type emptyStructVal struct {
		Name string `form:"name"`
	}

	var actual emptyStructVal
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		err := form.Unmarshal(r, &actual)
		if err != nil {
			t.Fatalf("unexpected unmarshal error: %s", err)
		}
	}))
	defer server.Close()

	r, err := http.NewRequest(http.MethodPost, server.URL, nil)
	if err != nil {
		t.Fatalf("unexpected error creating request: %s", err)
	}
	form := make(url.Values)
	form.Add("age", "30")
	r.URL.RawQuery = form.Encode()
	r.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(r)
	if err != nil {
		t.Fatalf("unexpected error sending request: %s", err)
	}
	resp.Body.Close()

	if actual.Name != "" {
		t.Fatalf("expected name to be empty. got=%s", actual.Name)
	}
}

func TestInvalidUnmarshalType(t *testing.T) {
	t.Parallel()
	type s struct {
		Val struct{} `form:"value"`
	}

	testUnmarshalFormError(t, "value", &s{}, "form: cannot unmarshal value into Go struct field s.Val of type struct {}: type struct {} cannot be unmarshalled from form")
}

func TestInvalidSliceTypeError(t *testing.T) {
	t.Parallel()
	type s struct {
		Val []int `form:"value"`
	}

	testUnmarshalFormError(t, "5,value", &s{}, "form: cannot unmarshal [5, value] into Go struct field s.Val of type []int: strconv.ParseInt: parsing \"value\": invalid syntax")
}

func TestInvalidArrayTypeError(t *testing.T) {
	t.Parallel()
	type s struct {
		Val [2]int `form:"value"`
	}

	testUnmarshalFormError(t, "5,6,value", &s{}, "form: cannot unmarshal [5, 6, value] into Go struct field s.Val of type [2]int: cannot use [3]int as [2]int value in struct")
	testUnmarshalFormError(t, "5,value", &s{}, "form: cannot unmarshal [5, value] into Go struct field s.Val of type [2]int: strconv.ParseInt: parsing \"value\": invalid syntax")
}

func TestMultipleFieldsForSingleValueError(t *testing.T) {
	t.Parallel()
	type s struct {
		Val int `form:"value"`
	}

	testUnmarshalFormError(t, "5,6", &s{}, "form: cannot unmarshal [5, 6] into Go struct field s.Val of type int: cannot unmarshal more than one value for non-slice field")
}

func testUnmarshalFormData[T constraints.Ordered](t *testing.T, expected UrlFormData[T]) {
	t.Helper()

	var actual UrlFormData[T]
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		err := form.Unmarshal(r, &actual)
		if err != nil {
			t.Fatalf("unexpected unmarshal error: %s", err)
		}
	}))
	defer server.Close()

	r, err := http.NewRequest(http.MethodPost, server.URL, nil)
	if err != nil {
		t.Fatalf("unexpected error creating request: %s", err)
	}
	form := make(url.Values)
	form.Add("single", fmt.Sprintf("%v", expected.Single))
	for _, val := range expected.Slice {
		form.Add("slice", fmt.Sprintf("%v", val))
	}
	for _, val := range expected.Array {
		form.Add("array", fmt.Sprintf("%v", val))
	}
	r.URL.RawQuery = form.Encode()
	r.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(r)
	if err != nil {
		t.Fatalf("unexpected error sending request: %s", err)
	}
	defer resp.Body.Close()

	if actual.Single != expected.Single {
		t.Fatalf("single value does not match. want=%v, got=%v", expected.Single, actual.Single)
	}
	if len(actual.Slice) != len(expected.Slice) {
		t.Fatalf("slices do not have the same length. want=%d, got=%d", len(expected.Slice), len(actual.Slice))
	}
	sort.Slice(actual.Slice, func(i, j int) bool {
		return actual.Slice[i] < actual.Slice[j]
	})
	sort.Slice(expected.Slice, func(i, j int) bool {
		return expected.Slice[i] < expected.Slice[j]
	})

	for i := 0; i < len(actual.Slice); i++ {
		if actual.Slice[i] != expected.Slice[i] {
			t.Fatalf("mismatch value in slice. want=%v, got=%v", expected.Slice[i], actual.Slice[i])
		}
	}

	sortArray(actual.Array)
	sortArray(expected.Array)

	for i := 0; i < len(actual.Slice); i++ {
		if actual.Array[i] != expected.Array[i] {
			t.Fatalf("mismatch value in array. want=%v, got=%v", expected.Array[i], actual.Array[i])
		}
	}
}

func TestParseFormError(t *testing.T) {
	t.Parallel()
	type s struct {
		Val int `form:"value"`
	}

	r, err := http.NewRequest(http.MethodPost, "/", nil)
	if err != nil {
		t.Fatalf("unexpected error creating request: %s", err)
	}

	err = form.Unmarshal(r, &s{})
	if err == nil {
		t.Fatalf("expected error from r.ParseForm()")
	}
}

func testUnmarshalFormError(t *testing.T, value string, i interface{}, expectedError string) {
	t.Helper()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		err := form.Unmarshal(r, i)
		if err == nil || err.Error() != expectedError {
			t.Errorf("wrong expected error. want=%s, got=%s", expectedError, err)
		}
	}))
	defer server.Close()

	r, err := http.NewRequest(http.MethodPost, server.URL, nil)
	if err != nil {
		t.Fatalf("unexpected error creating request: %s", err)
	}
	form := make(url.Values)
	for _, val := range strings.Split(value, ",") {
		form.Add("value", val)
	}
	r.URL.RawQuery = form.Encode()
	r.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(r)
	if err != nil {
		t.Fatalf("unexpected error sending request: %s", err)
	}
	resp.Body.Close()
}

func sortArray[T constraints.Ordered](a [2]T) {
	if a[0] > a[1] {
		temp := a[0]
		a[0] = a[1]
		a[1] = temp
	}
}
