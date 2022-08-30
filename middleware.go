package ratelimit

import (
	"encoding/json"
	"log"
	"net/http"
)

func Wrap(h http.Handler) *RateLimiterMiddleware {
	return &RateLimiterMiddleware{
		handler: h,
	}
}

type RateLimiterMiddleware struct {
	handler http.Handler
}

func (rl *RateLimiterMiddleware) fail(w http.ResponseWriter) {
	w.Header().Add("Content-Type", "application/json")

	w.WriteHeader(http.StatusTooManyRequests)

	r, _ := json.Marshal(map[string]string{
		"message": ErrTooManyRequests.Error(),
	})

	_, err := w.Write(r)
	if err != nil {
		panic(err)
	}
}

func (rl *RateLimiterMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ip := IP(r.RemoteAddr)

	if !Container.Has(r.Context(), ip) {
		err := Container.New(r.Context(), ip, Factory.Make(ip))
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)

			_, err = w.Write([]byte("Internal Server Error"))
			if err != nil {
				panic(err)
			}
			return
		}
	}

	if err := Container.Consume(r.Context(), ip); err != nil {
		if err == ErrTooManyRequests {
			rl.fail(w)
			return
		}
		log.Fatal(err)
	}

	rl.handler.ServeHTTP(w, r)
}
