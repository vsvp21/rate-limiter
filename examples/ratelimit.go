package examples

import (
	"github.com/vsvp21/rate-limiter"
	"log"
	"net/http"
	"time"
)

func finalHandler(w http.ResponseWriter, req *http.Request) {
	log.Print("Executing finalHandler")
	w.Write([]byte("OK"))
}

func main() {
	mux := http.NewServeMux()
	mux.Handle("/hello", http.HandlerFunc(finalHandler))

	rateLimited := ratelimit.Wrap(mux, 5, time.Second*5)

	err := http.ListenAndServe(":8081", rateLimited)
	if err != nil {
		log.Fatal(err)
	}
}
