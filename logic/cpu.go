package logic

import (
	"bytes"
	"context"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"math"
	"math/rand"
	"os"
	"os/exec"
	"strings"
)

const (
	PAGE_JUMP = 1024 * 4 // 4 kb the size of a page table
)

func DumpResults(result map[string]string, format string) {
	if format == "json" {
		result, err := json.MarshalIndent(result, "", "  ")
		if err != nil {
			fmt.Printf("Unable to marshall data: %s\n", err)
			os.Exit(1)
		}
		fmt.Println(string(result))
	} else if format == "print" {
		for key, val := range result {
			fmt.Println(key)
			fmt.Println(val)
		}
	} else {
		fmt.Printf("Unknown output format: %s\n", format)
		os.Exit(1)
	}
}

func PrintSchedulerStats(procType string, format string) {
	values := make(map[string]string, 20)
	// Set type of run in data for easier parsing
	values["type"] = procType
	var output bytes.Buffer
	procId := os.Getpid()
	cmd := exec.Command("cat", fmt.Sprintf("/proc/%d/schedstat", procId))
	cmd.Stdout = &output
	err := cmd.Run()
	if err != nil {
		fmt.Println("Failed get schedstat information")
		os.Exit(1)
	}
	schedstatVals := strings.Split(output.String(), " ")
	if len(schedstatVals) != 3 {
		fmt.Printf("Got unexpected number of values from schedstat: %d\n", len(schedstatVals))
		os.Exit(1)
	}
	result := output.String()
	// Doc Reference: https://github.com/torvalds/linux/blob/master/Documentation/scheduler/sched-stats.txt
	values["time_on_cpu"] = schedstatVals[0]
	values["wait_on_runqueue"] = schedstatVals[1]
	values["timeslices_ran"] = schedstatVals[2]
	output.Reset()
	cmd = exec.Command("cat", fmt.Sprintf("/proc/%d/stat", procId))
	cmd.Stdout = &output
	err = cmd.Run()
	if err != nil {
		fmt.Println("Failed get stat information")
		os.Exit(1)
	}
	statVals := strings.Split(output.String(), " ")
	values["utime_jiffies"] = statVals[13]
	values["kernal_time_jiffies"] = statVals[14]
	output.Reset()
	cmd = exec.Command("cat", fmt.Sprintf("/proc/%d/sched", procId))
	cmd.Stdout = &output
	err = cmd.Run()
	if err != nil {
		fmt.Println("Failed get sched information")
		os.Exit(1)
	}
	result = output.String()
	// Split the header from the body
	sections := strings.Split(result, "-\n")
	body := strings.Join(sections[1:], "")
	var subSections []string
	for _, section := range strings.Split(body, ":") {
		for _, line := range strings.Split(section, "\n") {
			subSections = append(subSections, strings.Trim(line, " "))
		}
	}
	hasKey := false
	var key string
	for _, section := range subSections {
		// Skips unnecessary fields at the end
		if strings.Contains(section, "=") {
			continue
		}
		if hasKey {
			values[key] = section
			hasKey = false
		} else {
			key = section
			hasKey = true
		}
	}
	DumpResults(values, format)
}

func QueryProc(procPath string) (string, error) {
	var output bytes.Buffer
	procId := os.Getpid()
	cmd := exec.Command("cat", fmt.Sprintf("/proc/%d/%s", procId, procPath))
	cmd.Stdout = &output
	err := cmd.Run()
	return output.String(), err
}

func PrintMemoryStats(procType string, format string) {
	values := make(map[string]string, 20)
	// Set type of run in data for easier parsing
	values["type"] = procType
	output, err := QueryProc("statm")
	if err != nil {
		fmt.Println("Failed get statm information")
		os.Exit(1)
	}
	schedstatVals := strings.Split(output, " ")
	if len(schedstatVals) != 6 {
		fmt.Printf("Got unexpected number of values from statm: %d\n", len(schedstatVals))
		os.Exit(1)
	}

	// Doc Reference: https://github.com/torvalds/linux/blob/master/Documentation/scheduler/sched-stats.txt
	values["total_size"] = schedstatVals[0]
	values["resident_set_size"] = schedstatVals[1]
	values["share"] = schedstatVals[2]

	output, err = QueryProc("status")
	if err != nil {
		fmt.Println("Failed get status information")
		os.Exit(1)
	}
	// Split the header from the body
	var subSections []string
	for _, section := range strings.Split(output, ":") {
		for _, line := range strings.Split(section, "\n") {
			subSections = append(subSections, strings.Trim(line, " "))
		}
	}
	hasKey := false
	var key string
	for _, section := range subSections {
		// Skips unnecessary fields at the end
		if strings.Contains(section, "=") {
			continue
		}
		if hasKey {
			values[key] = section
			hasKey = false
		} else {
			key = section
			hasKey = true
		}
	}
	if format == "json" {
		result, err := json.MarshalIndent(values, "", "  ")
		if err != nil {
			fmt.Printf("Unable to marshall data: %s\n", err)
			os.Exit(1)
		}
		fmt.Println(string(result))
	} else if format == "print" {
		for key, val := range values {
			fmt.Println(key)
			fmt.Println(val)
		}
	}
	DumpResults(values, format)
}

// Inidicate that the context is canceled
func IsCanceled(ctx context.Context) bool {
	for {
		select {
		case <-ctx.Done():
			return true
		default:
			return false
		}
	}
}

// Stole the logic from https://golang.org/doc/play/pi.go
func CPUIntensive(ctx context.Context) {
	var pi float64
	var counter uint64
	for {
		counter++
		pi += 4 * math.Pow(-1, float64(counter)) / float64((2*counter)+1)
		if IsCanceled(ctx) {
			break
		}
	}
}

// Basically the same as cpuIntensive, but writing the value
// of pi to a file and flushing it each time
func IOIntensive(ctx context.Context) {
	var pi float64
	var counter uint64
	tmp, err := ioutil.TempFile("", "io_intensive")
	if err != nil {
		panic(fmt.Sprintf("Failed to create temp file: %s", err))
	}
	defer os.Remove(tmp.Name())
	// Some data to write randomly to disk and then flush to disk
	dummyData := []byte("deadbeef")
	for {
		counter++
		pi += 4 * math.Pow(-1, float64(counter)) / float64((2*counter)+1)
		// Read then write from flushed file
		err = binary.Write(tmp, binary.BigEndian, &pi)
		if err != nil {
			panic(fmt.Sprintf("Failed to write: %s", err))
		}
		err = tmp.Sync()
		if err != nil {
			panic(fmt.Sprintf("Failed to sync: %s", err))
		}
		_, err = tmp.Seek(int64(-binary.Size(pi)), io.SeekEnd)
		if err != nil {
			panic(fmt.Sprintf("Failed to seek: %s", err))
		}
		_, err := tmp.Seek(4096*1024*5, io.SeekEnd)
		if err != nil {
			panic(fmt.Sprintf("Failed to do tail seek: %s", err))
		}
		offset, err := tmp.Seek(int64(os.SEEK_CUR), 0)
		if err != nil {
			panic(fmt.Sprintf("Failed to get current size: %s", err))
		}
		_, err = tmp.Seek(rand.Int63n(offset), io.SeekStart)
		if err != nil {
			panic(fmt.Sprintf("Failed to do random seek: %s", err))
		}
		tmp.Read(dummyData)
		tmp.Write(dummyData)
		err = tmp.Sync()
		if err != nil {
			panic(fmt.Sprintf("Failed to sync: %s", err))
		}
		if IsCanceled(ctx) {
			break
		}
	}
}

func EfficientMemoryUsage(toStore, format string) {
	output := make([]byte, len(toStore))
	for i := 0; i < len(toStore); i++ {
		output[i] = toStore[i]
	}
	PrintMemoryStats("efficient", format)
}

func InefficientMemoryUsage(toStore, format string) {
	output := make([]byte, len(toStore)+PAGE_JUMP*len(toStore))
	memoryIndex := 0
	for i := 0; i < len(toStore); i++ {
		output[memoryIndex] = toStore[i]
		memoryIndex += PAGE_JUMP
	}
	PrintMemoryStats("inefficient", format)
}
