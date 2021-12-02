package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/beebeeoii/lominus/pkg/api"
	"github.com/beebeeoii/lominus/pkg/auth"
)

func main() {
	jwtData, err := auth.LoadJwtData()

	_, fileFound := os.Stat(auth.CREDENTIALS_FILE_NAME) // checks if there is creds.gob

	if !(fileFound == nil) {
		var un string
		var pw string
		log.Println("creds.gob file not detected...")
		log.Println("Creating new creds.gob file...")
		log.Println("Enter your NUSNET username: ")
		fmt.Scanln(&un)
		log.Println("Enter your NUSNET password: ")
		fmt.Scanln(&pw)

		cred := auth.Credentials{Username: un, Password: pw}
		auth.SaveCredentials(cred)

	}

	if jwtData.IsExpired() {
		log.Println(&auth.JwtExpiredError{}) //seems to be causing some probs if Fatalln is used
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

	modReq := api.Request{
		Url:       api.MODULE_URL_ENDPOINT,
		JwtToken:  jwtData.JwtToken,
		UserAgent: api.USER_AGENT,
	}

	fmt.Println(modReq.GetModules())
	log.Printf("Time to expiry: %d hours", int(time.Until(time.Unix(jwtData.Expiry, 0)).Hours()))
}
