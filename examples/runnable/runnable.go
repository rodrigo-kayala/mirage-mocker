package main

import "net/http"

// Version returns current version runnable example
func Version(w http.ResponseWriter, r *http.Request, status int) error {
	version := "v1.0.0"

	w.Header().Add("Content-Type", "text/plain")
	w.WriteHeader(status)
	w.Write([]byte(version))

	return nil
}

func main() {

}
