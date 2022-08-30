package main

import (
	ratelimit "github.com/vsvp21/rate-limiter"
	"log"
	"net/http"
	"time"
)

func handler(w http.ResponseWriter, req *http.Request) {
	log.Print("Executing handler")
	w.Write([]byte("OK"))
}

func main() {
	mux := http.NewServeMux()
	mux.Handle("/hello", http.HandlerFunc(handler))

	ratelimit.Factory = ratelimit.NewTokenBucketStrategyFactory(1, time.Second*5)

	ratelimit.Container = ratelimit.NewRedisRateLimiterContainer("localhost", "6379", "password")

	rateLimited := ratelimit.Wrap(mux)

	err := http.ListenAndServe(":8081", rateLimited)
	if err != nil {
		log.Fatal(err)
	}
}
