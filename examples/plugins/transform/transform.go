package main

import "net/http"

// AddHeader adds an header to request
func AddHeader(r *http.Request) error {
	version := "v1.0.1"

	r.Header.Add("VERSION", version)

	return nil
}

func main() {

}
