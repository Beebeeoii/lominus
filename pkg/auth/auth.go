package auth

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
)

type constants struct {
	CodeUrl string
	JwtUrl  string
}

func Authenticate(username string, password string) {
	constFile, err := os.Open("./constants/auth.json")

	if err != nil {
		log.Fatalln(err)
	}

	defer constFile.Close()
	byteValue, _ := ioutil.ReadAll(constFile)

	var result map[string]interface{}
	json.Unmarshal([]byte(byteValue), &result)

	fmt.Println(result["jwtUrl"])
}
