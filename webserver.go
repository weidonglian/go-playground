package main

import (
	"fmt"
	"log"
	"net/http"
)

type Hello struct {
	Name string
}

func (h Hello) ServeHTTP(w http.ResponseWriter,	r *http.Request) {
	fmt.Fprint(w, "Hello, " + h.Name)
}

func main() {
	var h Hello = Hello{
		"Weidong",
	}
	if err := http.ListenAndServe("localhost:4000", h); err != nil {
		log.Fatal(h)
	}

}
