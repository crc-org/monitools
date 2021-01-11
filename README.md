monitools
==

Gather resource consumption data around a CodeReady Containers instance
  
## Standalone tool

### Install

Clone this repository by `gith clone https://github.com/jsliacan/monitools` and the following command will build the `monictl` executable and place it in `$(GOBIN)` directory.

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
2021/01/11 14:19:20 -------------
2021/01/11 14:19:20 Running monitoring tools with the following settings:
2021/01/11 14:19:20 Data directory: data
2021/01/11 14:19:20 Number of repeats: 5
2021/01/11 14:19:20 Pauses between repeats: 1s
2021/01/11 14:19:20 -------------
2021/01/11 14:19:20 going to record CPU usage percentage attributed to qemu
2021/01/11 14:19:20 going to retrieve crictl stats from the CRC VM
2021/01/11 14:19:26 recorded CPU usage percentage 5 times at 1 sec intervals
2021/01/11 14:19:26 crictl stats successfully retrieved
$ ls data
cpu.csv  crictl-stats-20210111150132.json
```

## 'Import into your code' usecase

First, install this module to your development environment using e.g. `go get` tool:

``` bash
go get github.com/jsliacan/monitools
```

Then use as any other package, e.g. by importing as:

``` bash
import github.com/jsliacan/monitools/tools
```
### Example

In `monitools/examples` you will find a short program that imports this module and the `tools` package and runs one of its functions. It assumes an existing CRC VM and probes for CPU usage of the `qemu` process 5 times with 1s sleep inbetween probes. Resulting %CPU is recorded in `cpu.csv` in the same `monitools/examples` folder. Run the example script `example.go` by 

``` bash
$ cd monitools/examples
$ go run example.go
```

