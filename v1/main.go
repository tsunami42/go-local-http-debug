package main

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/pprof"
	"os"
	"os/signal"
)

func logHandler(w http.ResponseWriter, r *http.Request) {
	// httpStart := time.Now()
	w.Header().Set("Content-Type", "text/plain")

	defer r.Body.Close()
	if body, err := ioutil.ReadAll(r.Body); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		io.WriteString(w, "Read from body failed")
	} else {
		// log.Println("Read time:", time.Since(httpStart))
		io.WriteString(w, fmt.Sprintf("body size %d", len(body)))
	}
}

func AttachProfiler(mux *http.ServeMux) {
	mux.HandleFunc("/debug/pprof/", pprof.Index)
	mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
	mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
}

func main() {
	mux := http.NewServeMux()
	AttachProfiler(mux)
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
