package main

import (
	// "bufio"
	LogInGo "LogInGo/pkg"
	"flag"

	// "fmt"
	"io"
	"log"
	"os"
	"strings"
)

func main() {
	out := flag.String("out", "stdout", "File name to use for log output. If stdout is provided, then output is written directly to the console.")
	// async := flag.Bool("async", false, "This flag determines if the logger should write asynchronously.")
	msgBufferSize := 1
	errorBufferSize := 1
	flag.Parse()

	var w io.Writer
	var err error
	if strings.ToLower(*out) == "stdout" {
		w = os.Stdout
	} else {
		w, err = os.Create(*out)
		if err != nil {
			log.Fatal("Unable to open log file", err)
		}
	}

	l := LogInGo.New(w, msgBufferSize, errorBufferSize)
	go l.Start()

	messageChan := l.MessageChannel()
	errChan := l.ErrorChannel()

	if errChan != nil {
		go func(errChan <-chan error) {
			err := <-errChan
			l.Stop()
			log.Fatalf("Error received from logger: %v\n", err)
		}(errChan)
	}

}
