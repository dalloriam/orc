package main

import (
	"fmt"

	"github.com/dalloriam/orc/orc/docker"
)

func main() {
	dock, err := docker.NewController(docker.Config{"./docker"})
	if err != nil {
		panic(err)
	}

	dock.Start("elasticsearch")

	fmt.Println("OK", dock)
}
