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

## Usage

### Quick Start

Build

```bash
task install
task protoc
task build
task run
```

### Server

You will generally want to start a server. You can either run it as a background
process or use `ps aux` from another terminal (or container with shared process namespace)
to see the PID.

```bash
./bin/server
```

To specify a different unix socket to use:

```bash
./bin/server -s /dinosaur.sock
```

See all options (there are few!)

```bash
./bin/server --help
```

If you leave out the socket, it by default will be written to '/tmp/goshare.sock'.
The same is true for the client, discussed next.

### Client

The client should connect to the same unix socket as the server for proper communication!

```bash
./bin/client -s /dinosaur.sock <command>
```

Since the jobs are run possibly on a different system, you can also define the working directory.

```bash
./bin/client -w /working/directory <command>
```

If you leave out a command, a dummy "echo hello world" is used for a test. 

### Wait

The wait executable is intended to take a command or full path to an executable, and return the PID. It will wait (and loop) until it finds a match.
As an example, here is with full verbosity:

```bash
$ ./bin/wait -c "/usr/libexec/gvfsd-recent --spawner :1.2 /org/gtk/gvfs/exec_spaw/5" 
```
```console
🟧️  wait: 2023/07/27 23:44:30 wait.go:38: /usr/libexec/gvfsd-recent --spawner :1.2 /org/gtk/gvfs/exec_spaw/5
🟧️  wait: 2023/07/27 23:44:30 wait.go:64: Found matched command /usr/libexec/gvfsd-recent --spawner :1.2 /org/gtk/gvfs/exec_spaw/5 with pid 4013899
4013899
```

We can run in quiet mode (only print out the final PID):

```bash
./bin/wait -c "/usr/libexec/gvfsd-recent --spawner :1.2 /org/gtk/gvfs/exec_spaw/5" -q
4013899
```

Or change the frequency of polling (in seconds):

```bash
./bin/wait -c "/usr/libexec/gvfsd-recent --spawner :1.2 /org/gtk/gvfs/exec_spaw/5" -q -w 10
4013899
```

Note that you can also provide the executable (full path) to look for if it is unique, and both should
be in quotes in case of flags, etc.

### Wait-fs

`wait-fs` is a derivative of wait, but will wait for a path on the filesystem to exist.
E.g., try running it with a path that doesn't exist in one terminal:

```bash
./bin/wait-fs -p ./does-not-exist.txt
```
```console
🟧️  wait-fs: 2023/09/06 13:50:04 wait-fs.go:53: Path ./does-not-exist.txt does not exist yet, sleeping 5
🟧️  wait-fs: 2023/09/06 13:50:09 wait-fs.go:53: Path ./does-not-exist.txt does not exist yet, sleeping 5
```

And then try creating it in another:

```bash
$ touch does-not-exist.txt
```

You'll see it discovered and the script exit.

```console
🟧️  wait-fs: 2023/09/06 13:50:04 wait-fs.go:40: ./does-not-exist.txt
🟧️  wait-fs: 2023/09/06 13:50:04 wait-fs.go:53: Path ./does-not-exist.txt does not exist yet, sleeping 5
🟧️  wait-fs: 2023/09/06 13:50:09 wait-fs.go:53: Path ./does-not-exist.txt does not exist yet, sleeping 5
🟧️  wait-fs: 2023/09/06 13:50:14 wait-fs.go:49: Found existing path ./does-not-exist.txt
```

And that's it! We use this for an operator where we are waiting for a particular file to exist.

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
task install
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
```console
task: [build] GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o bin/wait cmd/wait/wait.go
task: [build] GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o bin/server cmd/server/server.go
task: [build] GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o bin/client cmd/client/client.go
task: [build] GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o bin/wait-fs cmd/wait-fs/wait-fs.go
```

These are generated in [bin](bin)

### Run Server and Client

```sh
task run
```
```console
task: [run] ./bin/server &
task: [run] sleep 1
task: [run] ./bin/client
🟪️  client: 2023/07/25 15:57:16 client.go:40: socket path: /tmp/echo.sock
🟪️  client: 2023/07/25 15:57:16 client.go:41: requested command: echo hello world
🟪️  client: 2023/07/25 15:57:16 client.go:82: sent command: echo hello world
🟦️ service: 2023/07/25 15:57:16 command.go:26: start new stream request
🟦️ service: 2023/07/25 15:57:16 command.go:54: Received command echo hello world
🟦️ service: 2023/07/25 15:57:16 command.go:67: send new pid=461564
🟦️ service: 2023/07/25 15:57:16 command.go:70: Process started with PID: 461564
🟦️ service: 2023/07/25 15:57:16 command.go:75: send final output: hello world
🟪️  client: 2023/07/25 15:57:16 client.go:103: pid 461564 is active
🟪️  client: 2023/07/25 15:57:16 client.go:88: closing send
🟪️  client: 2023/07/25 15:57:16 client.go:103: pid 461564 is active
🟪️  client: 2023/07/25 15:57:16 client.go:107: new output received: hello world
🟪️  client: 2023/07/25 15:57:16 client.go:108: process is done, closing
🟪️  client: 2023/07/25 15:57:16 client.go:130: finished with client request
```

## TODO next

- add subcommands to client to run / cancel?
- ensure we check for executable first
 - should be table of values that indicate what happened
- test run with a sleep command, then cancel
- try making a release we can install to a dummy jobset with a flux container and go + application

## References

### Examples of both server and client side

 - https://github.com/devlights/go-grpc-uds-example
 - https://github.com/pahanini/go-grpc-bidirectional-streaming-example
 - https://zenn.dev/hsaki/books/golang-grpc-starting/viewer/client
 - http://yamahiro0518.hatenablog.com/entry/2016/02/01/215908
 - https://stackoverflow.com/a/46279623
 - https://stackoverflow.com/a/18479916

### gRPC with Unix Domain Socket example (server side)

 - https://qiita.com/hnakamur/items/848097aad846d40ae84b

### gRPC with Unix Domain Socket example (client side)

 - https://qiita.com/marnie_ms4/items/4582a1a0db363fe246f3

## License

HPCIC DevTools is distributed under the terms of the MIT license.
All new contributions must be made under this license.

See [LICENSE](https://github.com/converged-computing/cloud-select/blob/main/LICENSE),
[COPYRIGHT](https://github.com/converged-computing/cloud-select/blob/main/COPYRIGHT), and
[NOTICE](https://github.com/converged-computing/cloud-select/blob/main/NOTICE) for details.

SPDX-License-Identifier: (MIT)

LLNL-CODE- 842614
