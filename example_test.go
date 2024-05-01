package form_test

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"net/url"

	"github.com/hunterwilkins2/form"
)

type Person struct {
	Name string     `form:"name"`
	Age  int        `form:"age"`
	Pets []string   `form:"pets"`
	Nums [2]float32 `form:"nums"`
}

func ExampleUnmarshal() {
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

type Page struct {
	Page     int `form:"pageNumber"`
	PageSize int `form:"pageSize"`
}

func ExampleMarshal() {
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
