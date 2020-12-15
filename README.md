monitools
==

Performance monitoring tools for CRC.

## Types of data

1. Host side
   - Static
       - OS info in `/etc/os-release` or similar
	   - RAM
	   - Disk space
	   - \# CPUs
	   - hypervisor (incl. version)
	   - CRC version
	   - Openshift version
   - Dynamic
	   - CPU consumption attributed to `qemu` process
	   - Memory allocation for `qemu` process (constant)
   
2. VM side
   - Static
	   - OS info
	   - ???
   - CPU consumption ([in %]?)
   - Memory consumption (used [in %], free [in %]?)
   - Disk space (free [in %]?)

3. Cluster side 
   - `crictl stats` [per container]
   
## Ways of collecting

1. CPU consumption: poll `top` utility `t times` every `n sec`.
   - compute a moving average as variance is high

2. Check memory allocation for `qemu` process from `top` output?

3. Save output to file for `crictl stats -o yaml/json` command.

