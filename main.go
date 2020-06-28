package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/mkbblr/orch/internal/orch"

)

func main() {

	if len(os.Args) < 2 {
		log.Print("Usage: orch input.json")
		return
	}

	inFileName := os.Args[1]

	if _, err := os.Stat(inFileName); os.IsNotExist(err) {
		log.Print(err)
		return
	}

	inFile, err := os.Open(inFileName)
	if err != nil {
		log.Print(err)
		return
	}

	inBytes, err := ioutil.ReadAll(inFile)
	if err != nil {
		log.Print(err)
		return
	}

	//Bug: json.Unmarshal does not report error if json has duplicate keys.
	//Instead, it will merge and overwrite in the target map.
	//For the time being it is a burden user shall carry.
	var input map[string]interface{}
	err = json.Unmarshal(inBytes, &input)
	if err != nil {
		log.Print(err)
		return
	}

	orch.Start(input)

	fmt.Println("Done !!!")
}
