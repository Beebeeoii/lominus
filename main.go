package main

import (
	"log"
	"time"

	"github.com/beebeeoii/lominus/pkg/auth"
)

func main() {
	jwtData, err := auth.LoadJwtData()

	if jwtData.IsExpired() {
		log.Fatalln(&auth.JwtExpiredError{})
		log.Println("Retrieving new JWT token...")

		credentials, err := auth.LoadCredentials()
		if err != nil {
			log.Fatalln(err)
			return
		}
		_, err = auth.RetrieveJwtToken(credentials, true)
		if err != nil {
			log.Fatalln(err)
		}

		log.Println("Retrieved successfully.")
		jwtData, err = auth.LoadJwtData()
	}

	if err != nil {
		log.Fatalln(err)
		return
	}

	log.Printf("Time to expiry: %d hours", int(time.Until(time.Unix(jwtData.Expiry, 0)).Hours()))
}
