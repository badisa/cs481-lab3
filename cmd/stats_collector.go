package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/cs481-lab2/logic"
	"os"
	"os/exec"
	"strings"
	"sync"
	"syscall"
	"time"
)

const (
	PHASE_ALL   = "all"
	PHASE_CPU   = "cpu"
	PHASE_IO    = "io"
	PHASE_MIXED = "mixed"
	IO_CPU      = "io-cpu"
	PAGE_TABLE  = "page-table"
)

type StatsData struct {
	lock sync.RWMutex
	data map[string][]map[string]string
}

type InputData struct {
	mode      string
	phase     string
	maxTime   int
	stepSize  int
	procs     int
	batchSize int
}

func main() {
	mode := flag.String("mode", "cpu", "Whether to run the 'io-cpu' experiment or the 'page-table' experiment")
	phase := flag.String("phases", PHASE_ALL, "Type of phases to run. Options are all, cpu, io, mixed")
	maxTime := flag.Int("max-time", 5, "Maximum amount of time to run each proc")
	stepSize := flag.Int("step", 1, "How much to increase time step by until max is reached, also the starting time. In Seconds")
	procs := flag.Int("procs", 10, "How many procs to run in each iteration")
	batchSize := flag.Int("batch", 10, "Maximum number of processes to run simultaneously")
	flag.Parse()
	inputData := InputData{
		mode:      *mode,
		phase:     *phase,
		maxTime:   *maxTime,
		stepSize:  *stepSize,
		procs:     *procs,
		batchSize: *batchSize,
	}
	cmd := exec.Command("make", "build")
	var output bytes.Buffer
	cmd.Stdout = &output
	cmd.Stderr = &output
	err := cmd.Run()
	if err != nil {
		fmt.Printf("Failed to build: %s\n", output.String())
		os.Exit(1)
	}
	if _, err := os.Stat("pthread"); os.IsNotExist(err) {
		fmt.Printf("Failed to generate pthread binary")
		os.Exit(1)
	}
	defer os.Remove("pthread")
	runProcesses(&inputData)
}

func runProcesses(input *InputData) {
	var wg sync.WaitGroup
	stat := StatsData{data: make(map[string][]map[string]string, 10)}
	for runTime := input.stepSize; runTime <= input.maxTime; runTime += input.stepSize {
		fmt.Printf("Running Procs for %ds in batches of %d\n", runTime, input.batchSize)
		for i := 0; i < input.procs; i++ {
			if i%input.batchSize == 0 {
				wg.Wait()
			}
			wg.Add(1)
			go stat.RunProcess(input.mode, runTime, &wg)
		}
		wg.Wait()
	}
	path := fmt.Sprintf("lab-3-max-%d-step-%d-proc-%d-batches-%d.json", input.maxTime, input.stepSize, input.procs, input.batchSize)
	err := stat.Dump(path)
	if err != nil {
		fmt.Printf("Failed to dump: %s\n", err)
		os.Exit(1)
	}
	fmt.Printf("Wrote results to %s\n", path)
}

func (d *StatsData) RunProcess(mode string, maxTime int, wg *sync.WaitGroup) {
	defer wg.Done()
	cmd := exec.Command("./pthread", mode)
	var output bytes.Buffer
	cmd.Stdout = &output
	cmd.Stderr = &output
	err := cmd.Start()
	if err != nil {
		fmt.Printf("Failed to start for mode %s: error:%s\n", mode, err)
		return
	}
	proc := cmd.Process
	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(time.Second*time.Duration(maxTime)))
	defer cancel()
	for !logic.IsCanceled(ctx) {
		time.Sleep(time.Second)
	}
	err = proc.Signal(syscall.SIGQUIT)
	if err != nil {
		fmt.Printf("Failed to signal kill: %s\n", err)
		return
	}
	err = cmd.Wait()
	if err != nil {
		fmt.Printf("Failed to wait: %s\n", err)
		return
	}
	fmt.Println(output.String())
	fmt.Println(cmd.ProcessState.String())
	if !cmd.ProcessState.Success() {
		fmt.Printf("Failed to complete successfully: %t\n", cmd.ProcessState.Success())
		return
	}
	values := make(map[string]string, 20)
	results := strings.Split(output.String(), "::::::::::\n")
	schedStatVals := logic.ParseSchedStat(results[0])
	for key, val := range schedStatVals {
		values[key] = val
	}
	statVals := logic.ParseStat(results[1])
	for key, val := range statVals {
		values[key] = val
	}
	schedVals := logic.ParseSched(results[2])
	for key, val := range schedVals {
		values[key] = val
	}
	d.WriteRun(mode, values)
}

func (d *StatsData) WriteRun(key string, data map[string]string) {
	d.lock.Lock()
	defer d.lock.Unlock()
	if _, ok := d.data[key]; !ok {
		d.data[key] = []map[string]string{data}
	} else {
		d.data[key] = append(d.data[key], data)
	}
}

func (d StatsData) Dump(path string) error {
	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create output file:%s", err)
	}
	data, err := json.MarshalIndent(d.data, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to dump marshal data: %s", err)
	}
	_, err = file.Write(data)
	if err != nil {
		return fmt.Errorf("failed to write data to file: %s", err)
	}
	if err = file.Sync(); err != nil {
		return fmt.Errorf("failed to sync file: %s", err)
	}
	return nil
}
