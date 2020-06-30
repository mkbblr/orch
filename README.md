# orch
A basic command orchestrator. It runs the commands concurrently as specified in the input json file. Each command is executed in its own shell. So any changes by one command to the environment will not be available to the next command. 

## Build
This package is built with go1.4. Hopefully it should work with lower versions as well. However, that is not verified.
```
git clone https://github.com/mkbblr/orch
cd orch
GO111MODULE=on go build
```

## Run
```
./orch test/data/simple1.json
```

## Input
The input json should have the following format:
```
{
    "cmd-key-1": {
        "run": "shell-command-to-run"
        "dep": {
            "cmd-key-2": "pass|fail|end|start"
        }
    },
    "cmd-key-2": {
        "run": "shell-command-to-run"
    },
    ...
}
```
Refer [this](https://github.com/mkbblr/orch/blob/master/test/data/simple2.json) sample input file.


## Design
There are two packages - main, orch. orch is an internal package. The main function just read the input json and unmarshall it and pass on to orch.Start. Which in turn initialize and configure command objects and throw them out in the arena; with a sync.WaitGroup keeping vigil.

Each command object implement a observer and notifier interface to communicate change of state between dependent commands (observer pattern in action).

While configuring dependencies, it bails out if any dependency is missing. Post configuration, it checks for cyclic dependency and bails out if there is any. 


## TODO

* Unit tests
