package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/dalloriam/orc/interfaces"
)

func main() {
	cfg, err := LoadConfiguration()
	if err != nil {
		panic(err)
	}

	orcService, err := New(cfg, interfaces.HandleWithHTTP)
	if err != nil {
		panic(err)
	}
	fmt.Println(orcService)
	log.Fatal(http.ListenAndServe("0.0.0.0:8080", nil))
}
