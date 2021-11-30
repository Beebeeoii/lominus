package main

import (
	"fmt"
	"log"

	"github.com/beebeeoii/lominus/pkg/auth"
)

func main() {
	jwtToken, err := auth.Authenticate("nusstu\\eXXXXXXX", "p4ssw0rd")
	if err != nil {
		log.Fatalln(err)
	}

	fmt.Println(jwtToken)
}
