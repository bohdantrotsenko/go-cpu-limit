package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"sync/atomic"
	"time"
)

var addr = flag.String("addr", ":8787", "listening address")

const N = int64(1000)

func f() int64 {
	sum := int64(0)
	for i := int64(0); i < N; i++ {
		for j := int64(0); j < N; j++ {
			for k := int64(0); k < N; k++ {
				sum++
			}
		}
	}
	return sum
}

type srv struct {
	counter int32
}

func (s *srv) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	id := atomic.AddInt32(&s.counter, 1)
	defer atomic.AddInt32(&s.counter, -1)
	if r.RequestURI == "/status" {
		http.Error(w, "All fine", http.StatusOK)
		return
	}

	if r.RequestURI == "/load" {
		log.Printf("[id = %d] starting f()\n", id)
		before := time.Now()
		_ = f()
		duration := time.Now().Sub(before)
		log.Printf("[id = %d] run f() in %s, total: %d requests\n", id, duration, atomic.LoadInt32(&s.counter))
		http.Error(w, duration.String(), http.StatusOK)
		return
	}

	http.Error(w, "Not found", http.StatusNotFound)
}

func run() error {
	lis, err := net.Listen("tcp", *addr)
	if err != nil {
		return fmt.Errorf("listen %s", err)
	}
	log.Printf("Listening on %s, pid = %d\n", *addr, os.Getpid())

	service := &srv{}
	httpSrv := &http.Server{Handler: service}

	return httpSrv.Serve(lis)
}

func main() {
	defer log.Println("exited")
	flag.Parse()
	err := run()
	if err != nil {
		log.Fatal(err)
	}
}
