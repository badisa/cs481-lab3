package main

import (
	"flag"
	"github.com/cs481-lab2/logic"
)

func main() {
	messageToStore := flag.String("input", "deadbeef", "Message to store in memory")
	fmt := flag.String("format", "json", "Whether to print json or regular output (\"json\", \"print\")")
	flag.Parse()
	logic.EfficientMemoryUsage(*messageToStore, *fmt)
}
