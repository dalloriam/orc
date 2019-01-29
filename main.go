package main

import (
	"fmt"
	"net/http"

	"github.com/dalloriam/orc/orc"
)

func main() {
	orc, err := orc.New("./plugins")
	if err != nil {
		panic(err)
	}
	fmt.Println(orc)

	http.ListenAndServe("127.0.0.1:8080", nil)
}
