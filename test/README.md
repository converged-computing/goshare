# Testing Communication between Containers

We are going to test running this application in the context of a [shared process namespace between containers in a pod](https://kubernetes.io/docs/tasks/configure-pod-container/share-process-namespace/).

## Go Experiment

Create a cluster, and install JobSet:

```bash
kind create cluster
VERSION=v0.2.0
kubectl apply --server-side -f https://github.com/kubernetes-sigs/jobset/releases/download/$VERSION/manifests.yaml
```

And apply:

```bash
$ kubectl apply -f go.yaml
```

We will test this interactively for now. In the future we will want to:

- install the client/server depending on container
- find the correct PID for the running server based on matching some name or similar
- start the client with the common socket path.

### Server Container

In the server container, download the release for the server:

```bash
$ kubectl exec -it goshare-workers-0-0-jtxbr -c server bash
$ wget https://github.com/converged-computing/goshare/releases/download/2023-07/server
$ chmod +x ./server
$ mv ./server /bin
```

Start detached to see the process ID (there are other ways to do this eventually).
We will generate the socket at the root of the filesystem, and this is going to appear
at `/proc/<pid>/root` for the other container.

```bash
server -s /dinosaur.sock &
[1] 41
```

### Client Container

In a different terminal, do the same, but start the client.

```bash
$ kubectl exec -it goshare-workers-0-0-jtxbr -c client bash
$ wget https://github.com/converged-computing/goshare/releases/download/2023-07/client
$ chmod +x ./client
$ mv ./client /bin
```

And then we can verify the pid namespace is shared:

```bash
$ ls /proc/41/root
```
```console
bin   dev            etc  home  lib32  libx32  mnt  proc          product_uuid  run   srv  tmp  var
boot  dinosaur.sock  go   lib   lib64  media   opt  product_name  root          sbin  sys  usr
```

See "dinosaur.sock" above! Let's now connect to it, and try running different commands.
Each will hang until it's completed, giving us first back a PID and then output (if relevant):

```bash
client -s /proc/41/root/dinosaur.sock echo hello world
```
```console
🟪️  client: 2023/07/25 22:58:08 client.go:40: socket path: /proc/41/root/dinosaur.sock
🟪️  client: 2023/07/25 22:58:08 client.go:41: requested command: echo hello world
🟪️  client: 2023/07/25 22:58:08 client.go:82: sent command: echo hello world
🟪️  client: 2023/07/25 22:58:08 client.go:103: pid 66 is active
🟪️  client: 2023/07/25 22:58:08 client.go:88: closing send
🟪️  client: 2023/07/25 22:58:08 client.go:103: pid 66 is active
🟪️  client: 2023/07/25 22:58:08 client.go:107: new output received: hello world
🟪️  client: 2023/07/25 22:58:08 client.go:108: process is done, closing
🟪️  client: 2023/07/25 22:58:08 client.go:130: finished with client request
```
Note that the server continues running, but we see output!

```console
🟦️ service: 2023/07/25 22:58:08 command.go:26: start new stream request
🟦️ service: 2023/07/25 22:58:08 command.go:54: Received command echo hello world
🟦️ service: 2023/07/25 22:58:08 command.go:67: send new pid=66
🟦️ service: 2023/07/25 22:58:08 command.go:70: Process started with PID: 66
🟦️ service: 2023/07/25 22:58:08 command.go:75: send final output: hello world
```

We can try running another command (that will hang a bit more as it waits)

```bash
client -s /proc/41/root/dinosaur.sock sleep 10
```
```console
🟪️  client: 2023/07/25 22:59:11 client.go:40: socket path: /proc/41/root/dinosaur.sock
🟪️  client: 2023/07/25 22:59:11 client.go:41: requested command: sleep 10
🟪️  client: 2023/07/25 22:59:11 client.go:82: sent command: sleep 10
🟪️  client: 2023/07/25 22:59:11 client.go:103: pid 73 is active
🟪️  client: 2023/07/25 22:59:11 client.go:88: closing send
# Note there was a delay / wait here while the command was running
🟪️  client: 2023/07/25 22:59:21 client.go:103: pid 73 is active
🟪️  client: 2023/07/25 22:59:21 client.go:108: process is done, closing
🟪️  client: 2023/07/25 22:59:21 client.go:130: finished with client request
```

And the server also updated.

```console
🟦️ service: 2023/07/25 22:59:11 command.go:26: start new stream request
🟦️ service: 2023/07/25 22:59:11 command.go:54: Received command sleep 10
🟦️ service: 2023/07/25 22:59:11 command.go:67: send new pid=73
🟦️ service: 2023/07/25 22:59:11 command.go:70: Process started with PID: 73
```

The main difference for the second run is that we don't see output.
And that's it for this demo! Next we likely want to get a basic example running with Flux,
and then figure out how to automate the original process to get the server PID. Likely
we can have the flux run command (that will issue a command to the client) wait until it sees
a process running with a particular name.  When you are done, exit and clean up.

```bash
$ kubectl delete -f flux.yaml 
```