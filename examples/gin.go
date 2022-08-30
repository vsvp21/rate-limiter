package main

import (
	"github.com/gin-gonic/gin"
	ratelimit "github.com/vsvp21/rate-limiter"
	"log"
	"net/http"
)

func main() {
	r := gin.Default()
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})

	rateLimited := ratelimit.Wrap(r)

	err := http.ListenAndServe(":8081", rateLimited)
	if err != nil {
		log.Fatal(err)
		return
	}
}
