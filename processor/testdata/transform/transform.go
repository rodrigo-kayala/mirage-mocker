package main

import (
	"net/http"
	"os"
)

// AddHeader adds an header to request
func AddHeader(r *http.Request) error {
	vname, ok := r.URL.Query()["vname"]
	if ok && len(vname) > 0 {
		env := os.Getenv(vname[0])
		r.Header.Add(vname[0], env)
	}

	return nil
}
