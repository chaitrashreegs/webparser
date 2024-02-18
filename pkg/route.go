package pkg

import (
	"fmt"
	"net/http"
	"time"
)

// GetCounter returns a http.Handler that returns the total number of requests
func GetCounter(c Parser) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c.SendInput(time.Now())
		output := c.RecvOutput()
		fmt.Fprintf(w, "Total requests in the last %v: %d\n", c.GetWindowSize(), output)
	})
}
