package healthhandler

import "net/http"

func Get(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}
