monitools
==

Gather resource consumption data around a CodeReady Containers instance
  
## Container

The following diagram represent the interaction of monitools container with a host running the CodeReady Containers instance.

![Overview](docs/overview.jpg?raw=true)

The container will be configured with ENVs:

#### target host configuration   
  
**TARGET_HOST**(*mandatory*): Target host address   
**TARGET_HOST_USERNAME**(*mandatory*): username  

Chose one of following depending on the target auth mechanism:  

**TARGET_HOST_KEY_PATH**(*optional*): key path  
**TARGET_HOST_PASSWORD**(*optional*): pasword

#### monictl configuration  
  
**MONICTL_REPETITIONS**(*optional, default 5*)  
**MONICTL_INTERVAL**(*optional, default 1*)  

Sample container execution
```bash
podman run -it --rm \
        -e TARGET_HOST_USERNAME=XXXX \
        -e TARGET_HOST_KEY_PATH=XXXX \
        -e TARGET_HOST=XXXX \
 quay.io/crc/monitools
```

## Standalone tool

### Install

Clone this repository by `gith clone https://github.com/code-ready/monitools` and the following command will build the `monictl` executable and place it in `$(GOBIN)` directory.

``` bash
make install
```

### Usage

1. Create directory where you want `monictl` to deposit the data. 
2. Have a running instance of CRC (`crc status` returns `Running` for both VM and the cluster)
3. Run
   ```bash
   monictl -d=<data dir> -n=<num of reps> -s=<length of pause>
   ```
4. Look into your data folder to inspect the files.
   
### Flags/arguments

- `-d`: data directory (relative to current directory)
- `-n`: number of repetitions when probing CPU consumption
- `-s`: time interval between repetitions when probing CPU consumption (in seconds)

### Example of a first run

Assuming a running CRC instance:

``` bash
$ make install
$ mkdir data
$ monictl -d=data -n=5 -s=1
-------------
Running monitoring tools with the following settings:
Data directory: data
Number of repeats: 5
Pauses between repeats: 1s
Logging into: logs/monitools_20210329101528.log
-------------
$ ls data
cpu.json  crictl-stats-20210329101533.json  traffic.json
$ ls logs
monitools_20210329101528.log
```

## 'Import into your code' usecase

First, install this module to your development environment using e.g. `go get` tool:

``` bash
go get github.com/code-ready/monitools
```

Then use as any other package, e.g. by importing as:

``` bash
import github.com/code-ready/monitools/tools
```
### Example

In `monitools/examples` you will find a short program that imports this module and the `tools` package and runs one of its functions. It assumes an existing CRC VM and probes for CPU usage of the `qemu` process 5 times with 1s sleep inbetween probes. Resulting %CPU is recorded in `cpu.csv` in the same `monitools/examples` folder. Run the example script `example.go` by 

``` bash
$ cd monitools/examples
$ go run example.go
```

