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

## LAMMPS

This is an example of running LAMMPS across a pod. I'm still trying to figure out the right
use case for this, but the high level idea is that you can store your application logic (lammps)
separately from the workflow logic. The main difference with this example and the one above
is that we are using a pre-built version of the goshare binaries, and we use a config
map for custom entrypoints.

```bash
kind create cluster
VERSION=v0.2.0
kubectl apply --server-side -f https://github.com/kubernetes-sigs/jobset/releases/download/$VERSION/manifests.yaml
```

And apply the lammps config:

```bash
$ kubectl apply -f lammps.yaml
```

You can watch logs! In the client container (where we are issuing the command from, but does not have the application logic)
we can see the client request the command and get the output:

```bash
$ kubectl logs goshare-workers-0-0-zcv7n -c client -f
```
```console
Looking for PID for goshare-srv
Found PID 7 for goshare-srv
Running hello world
🟪️  client: 2023/07/26 23:55:32 client.go:42: socket path: /proc/7/root/dinosaur.sock
🟪️  client: 2023/07/26 23:55:32 client.go:43: requested command: echo hello world
🟪️  client: 2023/07/26 23:55:32 client.go:84: sent command: echo hello world
🟪️  client: 2023/07/26 23:55:32 client.go:105: pid 210 is active
🟪️  client: 2023/07/26 23:55:32 client.go:90: closing send
🟪️  client: 2023/07/26 23:55:32 client.go:105: pid 210 is active
🟪️  client: 2023/07/26 23:55:32 client.go:110: new output received: hello world
🟪️  client: 2023/07/26 23:55:32 client.go:112: process is done, closing
🟪️  client: 2023/07/26 23:55:32 client.go:134: finished with client request
Running lammps
🟪️  client: 2023/07/26 23:55:32 client.go:42: socket path: /proc/7/root/dinosaur.sock
🟪️  client: 2023/07/26 23:55:32 client.go:43: requested command: mpirun lmp -v x 1 -v y 1 -v z 1 -in in.reaxc.hns -nocite
🟪️  client: 2023/07/26 23:55:32 client.go:84: sent command: mpirun lmp -v x 1 -v y 1 -v z 1 -in in.reaxc.hns -nocite
🟪️  client: 2023/07/26 23:55:32 client.go:105: pid 218 is active
🟪️  client: 2023/07/26 23:55:33 client.go:90: closing send
🟪️  client: 2023/07/26 23:55:37 client.go:105: pid 218 is active
🟪️  client: 2023/07/26 23:55:37 client.go:110: new output received: LAMMPS (29 Sep 2021 - Update 2)
OMP_NUM_THREADS environment is not set. Defaulting to 1 thread. (src/comm.cpp:98)
  using 1 OpenMP thread(s) per MPI task
Reading data file ...
  triclinic box = (0.0000000 0.0000000 0.0000000) to (22.326000 11.141200 13.778966) with tilt (0.0000000 -5.0260300 0.0000000)
  1 by 1 by 1 MPI processor grid
  reading atoms ...
  304 atoms
  reading velocities ...
  304 velocities
  read_data CPU = 0.002 seconds
Replicating atoms ...
  triclinic box = (0.0000000 0.0000000 0.0000000) to (22.326000 11.141200 13.778966) with tilt (0.0000000 -5.0260300 0.0000000)
  1 by 1 by 1 MPI processor grid
  bounding box image = (0 -1 -1) to (0 1 1)
  bounding box extra memory = 0.03 MB
  average # of replicas added to proc = 1.00 out of 1 (100.00%)
  304 atoms
  replicate CPU = 0.001 seconds
Neighbor list info ...
  update every 20 steps, delay 0 steps, check no
  max neighbors/atom: 2000, page size: 100000
  master list distance cutoff = 11
  ghost atom cutoff = 11
  binsize = 5.5, bins = 5 3 3
  2 neighbor lists, perpetual/occasional/extra = 2 0 0
  (1) pair reax/c, perpetual
      attributes: half, newton off, ghost
      pair build: half/bin/newtoff/ghost
      stencil: full/ghost/bin/3d
      bin: standard
  (2) fix qeq/reax, perpetual, copy from (1)
      attributes: half, newton off, ghost
      pair build: copy
      stencil: none
      bin: none
Setting up Verlet run ...
  Unit style    : real
  Current step  : 0
  Time step     : 0.1
Per MPI rank memory allocation (min/avg/max) = 78.15 | 78.15 | 78.15 Mbytes
Step Temp PotEng Press E_vdwl E_coul Volume 
       0          300   -113.27833    427.09094   -111.57687   -1.7014647    3427.3584 
      10    298.13784   -113.27279    1855.1535   -111.57169   -1.7011017    3427.3584 
      20    294.02745   -113.25991    3911.5126     -111.559   -1.7009101    3427.3584 
      30    293.61692   -113.25867    7296.5076   -111.55793   -1.7007375    3427.3584 
      40    301.40293   -113.28175    9622.4058   -111.58127   -1.7004797    3427.3584 
      50    310.92489   -113.31003    10101.225   -111.60982   -1.7002117    3427.3584 
      60    311.37774   -113.31149    9274.1322   -111.61144   -1.7000446    3427.3584 
      70    302.58347   -113.28582     6350.705   -111.58587   -1.6999549    3427.3584 
      80    295.34242   -113.26406    6795.0642   -111.56427   -1.6997975    3427.3584 
      90    299.15724   -113.27518    9198.0327   -111.57566   -1.6995238    3427.3584 
     100    307.63997   -113.30058    9424.4991   -111.60129   -1.6992878    3427.3584 
Loop time of 3.58954 on 1 procs for 100 steps with 304 atoms

Performance: 0.241 ns/day, 99.709 hours/ns, 27.859 timesteps/s
99.9% CPU use with 1 MPI tasks x 1 OpenMP threads

MPI task timing breakdown:
Section |  min time  |  avg time  |  max time  |%varavg| %total
---------------------------------------------------------------
Pair    | 2.9354     | 2.9354     | 2.9354     |   0.0 | 81.78
Neigh   | 0.093552   | 0.093552   | 0.093552   |   0.0 |  2.61
Comm    | 0.0023255  | 0.0023255  | 0.0023255  |   0.0 |  0.06
Output  | 0.00020913 | 0.00020913 | 0.00020913 |   0.0 |  0.01
Modify  | 0.55761    | 0.55761    | 0.55761    |   0.0 | 15.53
Other   |            | 0.0004777  |            |       |  0.01

Nlocal:        304.000 ave         304 max         304 min
Histogram: 1 0 0 0 0 0 0 0 0 0
Nghost:        4443.00 ave        4443 max        4443 min
Histogram: 1 0 0 0 0 0 0 0 0 0
Neighs:        123880.0 ave      123880 max      123880 min
Histogram: 1 0 0 0 0 0 0 0 0 0

Total # of neighbors = 123880
Ave neighs/atom = 407.50000
Neighbor list builds = 5
Dangerous builds not checked
Total wall time: 0:00:03
🟪️  client: 2023/07/26 23:55:37 client.go:112: process is done, closing
🟪️  client: 2023/07/26 23:55:37 client.go:134: finished with client request
```
And in the server container, we can see that the request is done for lammps, we get the output, and of course it keeps running expecting
more commands.

```bash
$ kubectl logs goshare-workers-0-0-zcv7n -c client -f
```
```console
 service: 2023/07/26 23:55:06 server.go:38: starting service at socket /dinosaur.sock
🟦️ service: 2023/07/26 23:55:06 server.go:50: creating a new service to listen at /dinosaur.sock
🟦️ service: 2023/07/26 23:55:32 command.go:26: start new stream request
🟦️ service: 2023/07/26 23:55:32 command.go:54: Received command echo hello world
🟦️ service: 2023/07/26 23:55:32 command.go:67: send new pid=210
🟦️ service: 2023/07/26 23:55:32 command.go:70: Process started with PID: 210
🟦️ service: 2023/07/26 23:55:32 command.go:76: send final output: hello world
🟦️ service: 2023/07/26 23:55:32 command.go:26: start new stream request
🟦️ service: 2023/07/26 23:55:32 command.go:54: Received command mpirun lmp -v x 1 -v y 1 -v z 1 -in in.reaxc.hns -nocite
🟦️ service: 2023/07/26 23:55:32 command.go:67: send new pid=218
🟦️ service: 2023/07/26 23:55:32 command.go:70: Process started with PID: 218
🟦️ service: 2023/07/26 23:55:37 command.go:76: send final output: LAMMPS (29 Sep 2021 - Update 2)
OMP_NUM_THREADS environment is not set. Defaulting to 1 thread. (src/comm.cpp:98)
  using 1 OpenMP thread(s) per MPI task
Reading data file ...
  triclinic box = (0.0000000 0.0000000 0.0000000) to (22.326000 11.141200 13.778966) with tilt (0.0000000 -5.0260300 0.0000000)
  1 by 1 by 1 MPI processor grid
  reading atoms ...
  304 atoms
  reading velocities ...
  304 velocities
  read_data CPU = 0.002 seconds
Replicating atoms ...
  triclinic box = (0.0000000 0.0000000 0.0000000) to (22.326000 11.141200 13.778966) with tilt (0.0000000 -5.0260300 0.0000000)
  1 by 1 by 1 MPI processor grid
  bounding box image = (0 -1 -1) to (0 1 1)
  bounding box extra memory = 0.03 MB
  average # of replicas added to proc = 1.00 out of 1 (100.00%)
  304 atoms
  replicate CPU = 0.001 seconds
Neighbor list info ...
  update every 20 steps, delay 0 steps, check no
  max neighbors/atom: 2000, page size: 100000
  master list distance cutoff = 11
  ghost atom cutoff = 11
  binsize = 5.5, bins = 5 3 3
  2 neighbor lists, perpetual/occasional/extra = 2 0 0
  (1) pair reax/c, perpetual
      attributes: half, newton off, ghost
      pair build: half/bin/newtoff/ghost
      stencil: full/ghost/bin/3d
      bin: standard
  (2) fix qeq/reax, perpetual, copy from (1)
      attributes: half, newton off, ghost
      pair build: copy
      stencil: none
      bin: none
Setting up Verlet run ...
  Unit style    : real
  Current step  : 0
  Time step     : 0.1
Per MPI rank memory allocation (min/avg/max) = 78.15 | 78.15 | 78.15 Mbytes
Step Temp PotEng Press E_vdwl E_coul Volume 
       0          300   -113.27833    427.09094   -111.57687   -1.7014647    3427.3584 
      10    298.13784   -113.27279    1855.1535   -111.57169   -1.7011017    3427.3584 
      20    294.02745   -113.25991    3911.5126     -111.559   -1.7009101    3427.3584 
      30    293.61692   -113.25867    7296.5076   -111.55793   -1.7007375    3427.3584 
      40    301.40293   -113.28175    9622.4058   -111.58127   -1.7004797    3427.3584 
      50    310.92489   -113.31003    10101.225   -111.60982   -1.7002117    3427.3584 
      60    311.37774   -113.31149    9274.1322   -111.61144   -1.7000446    3427.3584 
      70    302.58347   -113.28582     6350.705   -111.58587   -1.6999549    3427.3584 
      80    295.34242   -113.26406    6795.0642   -111.56427   -1.6997975    3427.3584 
      90    299.15724   -113.27518    9198.0327   -111.57566   -1.6995238    3427.3584 
     100    307.63997   -113.30058    9424.4991   -111.60129   -1.6992878    3427.3584 
Loop time of 3.58954 on 1 procs for 100 steps with 304 atoms

Performance: 0.241 ns/day, 99.709 hours/ns, 27.859 timesteps/s
99.9% CPU use with 1 MPI tasks x 1 OpenMP threads

MPI task timing breakdown:
Section |  min time  |  avg time  |  max time  |%varavg| %total
---------------------------------------------------------------
Pair    | 2.9354     | 2.9354     | 2.9354     |   0.0 | 81.78
Neigh   | 0.093552   | 0.093552   | 0.093552   |   0.0 |  2.61
Comm    | 0.0023255  | 0.0023255  | 0.0023255  |   0.0 |  0.06
Output  | 0.00020913 | 0.00020913 | 0.00020913 |   0.0 |  0.01
Modify  | 0.55761    | 0.55761    | 0.55761    |   0.0 | 15.53
Other   |            | 0.0004777  |            |       |  0.01

Nlocal:        304.000 ave         304 max         304 min
Histogram: 1 0 0 0 0 0 0 0 0 0
Nghost:        4443.00 ave        4443 max        4443 min
Histogram: 1 0 0 0 0 0 0 0 0 0
Neighs:        123880.0 ave      123880 max      123880 min
Histogram: 1 0 0 0 0 0 0 0 0 0

Total # of neighbors = 123880
Ave neighs/atom = 407.50000
Neighbor list builds = 5
Dangerous builds not checked
Total wall time: 0:00:03
```
In the case of jobset, we could indicate that the jobset is successful when the client finishes, and this way the entire thing
would clean up.