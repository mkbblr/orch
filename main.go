package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/mkbblr/orch/internal/orch"

	"github.com/yourbasic/graph"
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

	u := new(orch.UI)

	commands := make(map[string]*orch.Command)

	i := 0
	//Initialize commmands
	for k, v := range input {
		c := new(orch.Command)
		c.Init(k, v, i)
		c.Register(u)
		commands[k] = c
		i++
	}

	g := graph.New(len(commands))

	//Register command dependencies and check for missing and cyclic dependency
	for _, c := range commands {
		err = c.RegisterNotifications(commands, g)
		if err != nil {
			return
		}
	}

	//this works with numeric command name only
	if !graph.Acyclic(g) {
		cycle := graph.StrongComponents(g)
		fmt.Println(cycle)
		fmt.Println("cyclic dependency detected in input, please fix it")
		return
	}

	//Execute commands concurrently
	for _, c := range commands {
		c.Execute()
	}

	//Wait until all goroutines finish their job
	orch.Wg.Wait()

	fmt.Println("Done !!!")
}
