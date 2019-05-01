package main

import (
	"context"
	"flag"
	"github.com/cs481-lab2/logic"
	"time"
)

func main() {
	secondsToCompletion := flag.Int("time", 10, "How long to run computation in seconds")
	fmt := flag.String("format", "json", "Format to print string with (\"json\", \"print\")")
	flag.Parse()
	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(time.Second*time.Duration(*secondsToCompletion)))
	defer cancel()
	logic.IOIntensive(ctx)
	logic.PrintSchedulerStats("io", *fmt)
}
