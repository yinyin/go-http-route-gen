package main

import (
	"flag"
	"log"
	"net/http"
	"time"
)

func parseCommandParam() (httpAddr string) {
	flag.StringVar(&httpAddr, "listen", ":8080", "port and address to listen on")
	flag.Parse()
	return
}

func main() {
	httpAddr := parseCommandParam()
	log.Printf("INFO: listen on address: [%s]", httpAddr)
	h := &sampleHandler{}
	s := &http.Server{
		Addr:         httpAddr,
		Handler:      h,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}
	log.Fatal(s.ListenAndServe())
}
