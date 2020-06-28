# orch
A basic command orchestrator. It runs the commands concurrently as specified in the input json file. Each command is executed in its own shell.

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
        "dep": "commma-separated-list-of-cmd-keys"
    },
    "cmd-key-2": {
        "run": "shell-command-to-run"
        "dep": "commma-separated-list-of-cmd-keys"
    }
}
```
Refer [this](https://github.com/mkbblr/orch/blob/master/test/data/simple2.json) sample input file.
