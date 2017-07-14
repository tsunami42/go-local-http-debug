package main

import (
	"bufio"
	"context"
	"io"
	"log"
	"net/http"
	"net/http/pprof"
	"os"
	"os/signal"
	"time"
)

const maxScanTokenSize = 1024 * 1024

var scanBuf []byte

func logHandler(w http.ResponseWriter, r *http.Request) {
	httpStart := time.Now()
	w.Header().Set("Content-Type", "text/plain")

	defer r.Body.Close()

	s := bufio.NewScanner(r.Body)
	s.Buffer(scanBuf, maxScanTokenSize)

	for s.Scan() {
	}

	log.Println("Read time:", time.Since(httpStart))
	if s.Err() != nil {
		w.WriteHeader(http.StatusInternalServerError)
		io.WriteString(w, "Read from body failed")
	} else {
		io.WriteString(w, "Read Successful")
	}
}

func attachProfiler(mux *http.ServeMux) {
	mux.HandleFunc("/debug/pprof/", pprof.Index)
	mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
	mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
}

func main() {
	scanBuf = make([]byte, 100*1024)

	mux := http.NewServeMux()
	attachProfiler(mux)
	mux.HandleFunc("/log/", logHandler)
	server := http.Server{
		Addr:    ":8094",
		Handler: mux,
	}

	go func() {
		if err := server.ListenAndServe(); err != nil {
			log.Fatalln(err)
		}
	}()

	defer func() {
		log.Println("shutdown http")
		if err := server.Shutdown(context.Background()); err != nil {
			log.Fatalln(err)
		}
	}()

	signals := make(chan os.Signal)
	signal.Notify(signals, os.Interrupt)
Loop:
	select {
	case <-signals:
		log.Println("signal received")
		break Loop
	}
}
