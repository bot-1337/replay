# `replay`

`replay` is a CLI for determining the state of remote systems at a given point in time

## Development Requirements

Developing on this project requires `docker` + `docker-compose`. The project is written in golang, but a local golang installation is not required.

## Running

To run tests => `make test`

To build the CLI and run it, do the following

``` bash
$ make dev # <= get dev assets
$ make build # <= build the CLI
$ ./replay --field ambientTemp --field schedule /tmp/ehub_data 2016-01-01T03:00
$ ./replay --field ambientTemp --field schedule --debug /tmp/ehub_data 2016-01-01T03:00 # <= in debug mode
$ ./replay --field ambientTemp --field schedule s3://net.energyhub.assets/public/dev-exercises/audit-data/ 2016-01-01T03:00
$ ./replay --field ambientTemp --field schedule --debug s3://net.energyhub.assets/public/dev-exercises/audit-data/ 2016-01-01T03:00 # <= in debug mode
```

## Code Architecture

The code is setup as the following 3 significant layers:

{ `CLI` } = talks to the => { `Controller` } = talks to the => { `Reader` }

1. The `CLI` (in `cli.go`) layer does "front door" user input validation, and provides the framework for executing other code
2. The `Controller` (in `controller.go`) layer contains the primary business logic of the application
3. The `Reader` (in `reader.go`) layer contains the functions for reading from your local machine, and from s3

There is also the time parsing utility layer (`time.go`) which is necessary for all the time parsing in this problem space.
