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

	if err := dock.Start("elasticsearch"); err != nil {
		panic(err)
	}

	fmt.Println("OK", dock)
}
