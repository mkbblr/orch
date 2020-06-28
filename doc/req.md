## Requirements
Design and implement a simple command orchestrator as follows:

1. The commands should be input in a machine parsable format (JSON/XML).
2. The orchestrator should read the incoming command and execute them concurrently.
3. At any given time, the user should be able to see the current state of the orchestrator:
    - How many commands have been executed with completion status.
    - How many commands are in progress.
    - How many commands are pending.

4.  Some commands might have inter-dependencies that should be expressible in the input to the orchestrator.
    a.  The orchestrator should honour such dependencies.

Other considerations:

1.  Choose any high level programming language.
2.  Provide a README explaining the code layout and general design considerations.