package main

import (
	"errors"
	"flag"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
)

const usageFormat = `Demo of Dining Philosophers Problem (https://en.wikipedia.org/wiki/Dining_philosophers_problem)

usage: %s -http=%s
`

var addr *string

func init() {
	const defaultAddr = "localhost:8888"
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), usageFormat, filepath.Base(os.Args[0]), defaultAddr)
		flag.PrintDefaults()
	}
	addr = flag.String("http", defaultAddr, "HTTP service address")
	flag.Parse()
}

func main() {
	fmt.Fprintf(os.Stdout, "Starting HTTP server %v ...\n", *addr)
	err := StartServer(*addr)
	if !errors.Is(err, http.ErrServerClosed) {
		panic(err)
	}
}
