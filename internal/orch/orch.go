package orch

import (
	"errors"
	"fmt"
	"os/exec"
	"strings"
	"sync"

	"github.com/spf13/cast"
	"github.com/yourbasic/graph"
)

type (
	//Event ... event
	Event struct {
		key string
		evt string
		ret error
	}

	//Notifier ... notifier
	Notifier interface {
		Register(Observer)
		Deregister(Observer)
		Notify(Event)
	}

	//Observer ... observer
	Observer interface {
		OnNotify(Event)
	}

	//Command ... command
	Command struct {
		idx  int    //this is used by graph package for cycle check
		key  string //command key as per input file
		cmd  string //command content
		ret  error  //command execution return value
		done bool   //command execution completed or not

		observers map[Observer]interface{}
		trackdep  map[string]bool
	}

	//UI ... ui/console updater
	UI struct {
	}
)

//Init ... initialize a command from un-marshalled json content
func (c *Command) Init(k string, v interface{}, idx int) {
	c.idx = idx
	c.key = k

	m := v.(map[string]interface{})
	c.cmd = m["run"].(string)

	c.ret = nil
	c.done = false

	c.observers = make(map[Observer]interface{})
	c.trackdep = make(map[string]bool)
	if v, ok := m["dep"]; ok {
		deps := strings.Split(v.(string), ",")
		for _, d := range deps {
			c.trackdep[d] = false
		}
	}
}

//OnNotify ... process notification from other commands (dependencies)
func (c *Command) OnNotify(e Event) {
	// log.Print("Command " + c.key + " received " + e.evt + " of " + e.key)
	if e.evt == "end" {
		c.trackdep[e.key] = true
		c.execute()
	}
}

//Notify ... notify dependent commands
func (c *Command) Notify(e Event) {
	// log.Print("Command " + c.key + " notifying  command " + e.evt + "  to all observers")
	for o := range c.observers {
		o.OnNotify(e)
	}
}

//Register ... register a observer
func (c *Command) Register(o Observer) {
	c.observers[o] = nil
}

//Deregister ... deregister a observer
func (c *Command) Deregister(o Observer) {
	delete(c.observers, o)
}

//Execute ... execute a command after verifying all dependencies
func (c *Command) Execute() {
	wg.Add(1)
	c.execute()
}

func (c *Command) execute() {
	//check if all dependencies are satisfied
	for _, v := range c.trackdep {
		if !v {
			return
		}
	}

	c.Notify(Event{key: c.key, evt: "start"})
	go func() {
		defer wg.Done()
		execCmd := exec.Command("/bin/sh", "-c", c.cmd)
		c.ret = execCmd.Run()
		c.done = true
		// log.Print("++++++++++++++" + c.cmd)
		c.Notify(Event{key: c.key, evt: "end", ret: c.ret})
	}()
}

//RegisterNotifications ... register notifications from all dependencies
func (c *Command) RegisterNotifications(g *graph.Mutable) error {

	for d := range c.trackdep {
		if notifier, ok := commands[d]; ok {
			notifier.Register(c)
			if !g.Edge(c.idx, notifier.idx) {
				g.Add(c.idx, notifier.idx)
			}
		} else {
			fmt.Println(c.key + " depends upon " + d + ", which is missing !!!")
			fmt.Println("fix the input and run again")
			return errors.New("missing dependency [" + c.key + "] -> [" + d + "]")
		}
	}

	return nil
}

//OnNotify ... update ui on event notification
func (u *UI) OnNotify(e Event) {
	if e.evt == "end" {
		completedCount += 1
		fmt.Println( "command " + e.key+" completed, ret: ", e.ret, ",  " + cast.ToString(completedCount) + "/" + cast.ToString(len(commands)))
	}
}



var wg sync.WaitGroup
var commands map[string]*Command 
var completedCount int

func init() {
	commands = make(map[string]*Command)
}



func Start(input map[string]interface{}) {

	u := new(UI)


	//Initialize commands
	i := 0
	for k, v := range input {
		c := new(Command)
		c.Init(k, v, i)
		c.Register(u)
		commands[k] = c
		i++
	}

	g := graph.New(len(commands))

	//Register command dependencies and check for missing and cyclic dependency
	for _, c := range commands {
		err := c.RegisterNotifications(g)
		if err != nil {
			return
		}
	}

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
	wg.Wait()

}
