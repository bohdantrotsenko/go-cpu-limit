package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"sync/atomic"
	"time"
)

var target = flag.String("target", "http://localhost:8787/load", "target")
var requests = flag.Int("r", 80, "number of requests")

func load(counter *int32, errChan chan error) {
	defer func() {
		if atomic.AddInt32(counter, -1) == 0 {
			close(errChan)
		}
	}()

	before := time.Now()
	resp, err := http.Get(*target)
	if resp != nil && resp.Body != nil {
		defer resp.Body.Close()
	}
	if err != nil {
		errChan <- err
		return
	}
	reportedTime, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		errChan <- err
		return
	}

	duration := time.Now().Sub(before)

	fmt.Printf("actual: %s, reported: %s\n", duration.String(), strings.TrimSpace(string(reportedTime)))
}

func run() error {
	errChan := make(chan error, 1)

	counter := int32(*requests)
	for i := *requests; i > 0; i-- {
		go load(&counter, errChan)
	}

	for err := range errChan {
		return err
	}
	return nil
}

func main() {
	flag.Parse()
	err := run()
	if err != nil {
		log.Fatal(err)
	}
}
