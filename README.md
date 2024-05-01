# Go Form Marshaller/Unmarshaller

[![Go Reference](https://pkg.go.dev/badge/github.com/hunterwilkins2/form/slug.svg)](https://pkg.go.dev/github.com/hunterwilkins2/form)
![Unit tests](https://github.com/hunterwilkins2/form/actions/workflows/test.yaml/badge.svg)

Package form marshalles/unmarshalles Go structs into [*http.Request] forms

This package adds the struct tag "form", only fields with this tag will be marshalled/unmarshalled.
All primative types including their slice and array equivalent are supported.
Those include bool, string, int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64,
float32, float64, complex64, complex128.

## Installation

```
go get github.com/hunterwilkins2/form
```

## Example

```go
package form_test

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"net/url"

	"github.com/hunterwilkins2/form"
)

func ExampleUnmarshal() {
	type Person struct {
		Name string     `form:"name"`
		Age  int        `form:"age"`
		Pets []string   `form:"pets"`
		Nums [2]float32 `form:"nums"`
	}

	r, _ := http.NewRequest(http.MethodPost, "/users", bytes.NewBuffer([]byte{}))
	reqForm := make(url.Values)
	reqForm = url.Values{
		"name": []string{"John"},
		"age":  []string{"24"},
		"pets": []string{"Sam", "Spot", "Chester"},
		"nums": []string{"10", "20"},
	}
	r.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	r.URL.RawQuery = reqForm.Encode()

	p := &Person{}
	err := form.Unmarshal(r, p)
	if err != nil {
		log.Fatalf("Could not unmarshal person: %s", err)
	}

	fmt.Println(p)
	// Output: &{John 24 [Sam Spot Chester] [10 20]}
}

func ExampleMarshal() {
	type Page struct {
		Page     int `form:"pageNumber"`
		PageSize int `form:"pageSize"`
	}

	p := Page{
		Page:     2,
		PageSize: 200,
	}
	r, _ := http.NewRequest(http.MethodGet, "/products", nil)
	err := form.Marshal(r, &p)
	if err != nil {
		log.Fatalf("Could not marshal page query: %s", err)
	}

	fmt.Println(r.URL)
	// Output: /products?pageNumber=2&pageSize=200
}
```
