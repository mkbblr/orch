package orch

import (
	"errors"
	"fmt"
	"os/exec"
	"sync"

	"github.com/gookit/color"
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
		trackdep  map[string]Dependency
	}

	//Dependency ... an object to model a dependency
	Dependency struct {
		depType string //"pass|fail|start|end"
		ok      bool   //if met or not
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
	c.trackdep = make(map[string]Dependency)
	if v, ok := m["dep"]; ok {
		dep := v.(map[string]interface{})
		for k, v := range dep {
			c.trackdep[k] = Dependency{depType: v.(string), ok: false}
		}
	}
}

func (c *Command) onDepStart(e Event) {
	d, ok := c.trackdep[e.key]
	if !ok {
		return
	}

	if d.depType != "start" {
		return
	}

	c.trackdep[e.key] = Dependency{depType: d.depType, ok: true}
	c.execute()
}

func (c *Command) onDepEnd(e Event) {
	d, ok := c.trackdep[e.key]
	if !ok {
		return
	}

	if d.depType == "end" {
		c.trackdep[e.key] = Dependency{depType: d.depType, ok: true}
		c.execute()
		return
	}

	if e.ret == nil {
		c.onDepPass(e)
	} else {
		c.onDepFail(e)
	}
}
func (c *Command) onDepPass(e Event) {
	d, ok := c.trackdep[e.key]
	if !ok {
		return
	}

	if d.depType == "pass" {
		c.trackdep[e.key] = Dependency{depType: d.depType, ok: true}
		c.execute()
		return
	}

	if d.depType == "fail" {
		color.Error.Println("abort " + c.key + " due to " + e.key + " pass")

		c.abort()
		return
	}

}
func (c *Command) onDepFail(e Event) {
	d, ok := c.trackdep[e.key]
	if !ok {
		return
	}

	if d.depType == "fail" {
		c.trackdep[e.key] = Dependency{depType: d.depType, ok: true}
		c.execute()
	}

	if d.depType == "pass" {
		color.Error.Println("abort " + c.key + " due to " + e.key + " fail")
		c.abort()
		return
	}

}

func (c *Command) onDepAbort(e Event) {
	d, ok := c.trackdep[e.key]
	if !ok {
		return
	}

	if d.depType == "abort" || d.depType == "start" {
		c.trackdep[e.key] = Dependency{depType: d.depType, ok: true}
		c.execute()
	}

	c.abort()
}

//OnNotify ... process notification from other commands (dependencies)
func (c *Command) OnNotify(e Event) {

	switch {
	case e.evt == "start":
		c.onDepStart(e)
		break
	case e.evt == "end":
		c.onDepEnd(e)
		break
	case e.evt == "abort":
		c.onDepAbort(e)
		break
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
		if !v.ok {
			return
		}
	}

	c.Notify(Event{key: c.key, evt: "start", ret: nil})
	go func() {
		defer wg.Done()
		execCmd := exec.Command("/bin/sh", "-c", c.cmd)
		c.ret = execCmd.Run()
		c.done = true
		// log.Print("++++++++++++++" + c.cmd)
		c.Notify(Event{key: c.key, evt: "end", ret: c.ret})
	}()
}

func (c *Command) abort() {
	defer wg.Done()
	c.done = true
	c.ret = errors.New("aborted")
	c.Notify(Event{key: c.key, evt: "abort", ret: c.ret})
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

			color.Error.Println("\"" + c.key + "\" depends upon \"" + d + "\", which is missing !!!")
			fmt.Println("fix the input and run again")
			return errors.New("missing dependency [" + c.key + "] -> [" + d + "]")
		}
	}

	return nil
}

//OnNotify ... update ui on event notification
func (u *UI) OnNotify(e Event) {
	mtx.Lock()
	defer mtx.Unlock()
	if e.evt == "end" {
		completed++
		inProgress--

		if e.ret == nil {
			pass++
		} else {
			fail++
		}

		// fmt.Println("")
		// fmt.Println("==== end: " + e.key)
	} else if e.evt == "abort" {
		aborted++
		// fmt.Println("")
		// fmt.Println("**** abort: " + e.key)
	} else if e.evt == "start" {
		inProgress++
		// fmt.Println("")
		// fmt.Println("++++ start: " + e.key)
	}
	total := len(commands)
	pending := total-completed-inProgress-aborted

	txt := cast.ToString(completed) + "/" + cast.ToString(total) + " completed (" + cast.ToString(pass) + " pass, " + cast.ToString(fail) + " fail), " + cast.ToString(inProgress) + " in progress, " + cast.ToString(aborted) + " aborted, " + cast.ToString(pending) + " Pending."
	fmt.Println(txt)

}

var wg sync.WaitGroup
var commands map[string]*Command
var completed int
var pass int
var fail int
var inProgress int
var aborted int
var mtx sync.Mutex

func init() {
	commands = make(map[string]*Command)
}

//Start ... start orchestartor with given input
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
		// cycle := graph.StrongComponents(g)
		// fmt.Println(cycle)
		color.Error.Println("cyclic dependency detected in input, please fix it")
		return
	}

	//Execute commands concurrently
	for _, c := range commands {
		c.Execute()
	}

	//Wait until all goroutines finish their job
	wg.Wait()

}
