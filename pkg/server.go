package pkg

import (
	"net/http"
)

// NewServer creates a new server with the given URL and http.ServeMux.
func NewServer(url string, mx *http.ServeMux) error {
	err := http.ListenAndServe(url, mx)
	if err != nil {
		return err
	}
	return nil
}
