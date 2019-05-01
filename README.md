Lab 3 - UNM CS481
=================
---

Summary
-------
A lab from UNM's CS481 to the effect on pthreads when running IO and CPU intensive
tasks.

Components
----------

### `Golang` - Tested using 1.11
  - Scripts for generating results related to the CFS scheduler

### `Python` - Requires 3.4+
  - Jupyter Notebook for data analysis and graph generation


Installation
------------

### Golang

Refer to the golang [docs](https://golang.org/doc/install). 

### Python

#### TODO

Running
-------

### Golang

```
// To run processes for 8 to 64 seconds (incrementing by 8) with 100
// proccesses run concurrently in each phase (IO only, CPU only, Half and Half)
$ go run cmd/stats_collector.go -time 64 -step 8 -procs 100
```

### Python
The python component is a Ipython notebook, allowing review of how the analysis was performed.
```
// Start the notebook server, then navigate to the analysis.ipynb
$ jupyter notebook
```

