package main

import (
	"net/http"
	"os"
)

// GetEnv returns an envoriment variable
// DO NOT USE IT IN PRODUCTION!!!
func GetEnv(w http.ResponseWriter, r *http.Request, status int) error {
	var env string
	vname, ok := r.URL.Query()["vname"]
	if ok && len(vname) > 0 {
		env = os.Getenv(vname[0])
	}

	w.Header().Add("Content-Type", "text/plain")
	w.WriteHeader(status)
	w.Write([]byte(env))

	return nil
}
