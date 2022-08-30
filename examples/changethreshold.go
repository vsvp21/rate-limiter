package main

import (
	ratelimit "github.com/vsvp21/rate-limiter"
	"log"
	"net/http"
	"time"
)

func handler1(w http.ResponseWriter, req *http.Request) {
	log.Print("Executing handler")
	w.Write([]byte("OK"))
}

func main() {
	mux := http.NewServeMux()
	mux.Handle("/hello", http.HandlerFunc(handler1))

	ratelimit.Factory = ratelimit.NewTokenBucketStrategyFactory(1, time.Second*5)

	rateLimited := ratelimit.Wrap(mux)

	err := http.ListenAndServe(":8081", rateLimited)
	if err != nil {
		log.Fatal(err)
	}
}
