# goshare

Producer / consumer model to share commands between containers in Kubernetes. We do this using gRPC over unix domain sockets (UDS) via:

- Running a process in the consumer container with a listener. This creates a PID that the producer can find in `/proc/<pid>`
- Start the producer, pointing it to the PID of the consumer, and expecting it to write a socket to a known path in `/proc/<pid>/root`

At this point, we can run the producer as many times as needed, providing a command to give to the consumer to execute. The consumer will:

 - Receive the command
 - Execute it (or return an error back it's not found, etc.)
 - Provide the pid back to the producer
 - The producer then needs to somehow watch this PID for it to complete (likely with some API that uses ps, need to think about this more because we don't want to be polling)

I am starting from [this example](https://github.com/devlights/go-grpc-uds-example/tree/master) with an MIT license, included in [.github](.github).
I need creative terminology for producer and consumer, so I'm stil thinking about this. Right now, client and server is probably logical
for a listener and message sender! I'm first going to test this small app (to make sure it works) and then I'll work on customizing it
for submitting jobs. I am reading that we should set `GOMAXPROCS` to be the number of concurrent jobs we will allow.

## Setup

We are going to use [go-task](https://taskfile.dev/) over a Makefile. To install, [download a release](https://github.com/go-task/task/releases) and I installed with dpkg.

```sh
$ task --list
task: Available tasks for this project:
* build:                      build
* install-requirements:       install requirements
* protoc:                     gen protoc
* run:                        run
```

### Install gRPC and Go libraries

```sh
task install-requirements
```

### Run protoc

The way I understand this, this compiles the code from [proto](proto) (the echo.proto) into the [internal](internal) folder
where it can be used by the Go libraries under [cmd](cmd) to define the structure of messages.

```bash
task protoc
```

### Build Server and Client

```bash
task build
```

These are generated in [bin](bin)

### Run Server and Client

```sh
task run
```

## TODO next

- add background workers to handle waiting for task after run
- add more fields to proto for pid, retval (if it didn't work), and command?
- add subcommands to client to run / cancel
- ensure we check for executable first
 - should be table of values that indicate what happened
- test run with a sleep command, then cancel
- then need a way to get output / status based on pid without polling...
- try making a release we can install to a dummy jobset with a flux container and go + application

## References

### gRPC with Unix Domain Socket example (server side)

 - https://github.com/pahanini/go-grpc-bidirectional-streaming-example
 - https://qiita.com/marnie_ms4/items/4582a1a0db363fe246f3
 - http://yamahiro0518.hatenablog.com/entry/2016/02/01/215908
 - https://zenn.dev/hsaki/books/golang-grpc-starting/viewer/client
 - https://stackoverflow.com/a/46279623
 - https://stackoverflow.com/a/18479916
 - https://qiita.com/hnakamur/items/848097aad846d40ae84b

### gRPC with Unix Domain Socket example (client side)

 - https://github.com/pahanini/go-grpc-bidirectional-streaming-example
 - https://qiita.com/marnie_ms4/items/4582a1a0db363fe246f3
 - http://yamahiro0518.hatenablog.com/entry/2016/02/01/215908
 - https://zenn.dev/hsaki/books/golang-grpc-starting/viewer/client
 - https://stackoverflow.com/a/46279623
 - https://stackoverflow.com/a/18479916

## License

HPCIC DevTools is distributed under the terms of the MIT license.
All new contributions must be made under this license.

See [LICENSE](https://github.com/converged-computing/cloud-select/blob/main/LICENSE),
[COPYRIGHT](https://github.com/converged-computing/cloud-select/blob/main/COPYRIGHT), and
[NOTICE](https://github.com/converged-computing/cloud-select/blob/main/NOTICE) for details.

SPDX-License-Identifier: (MIT)

LLNL-CODE- 842614