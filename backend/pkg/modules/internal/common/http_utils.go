package common

import (
	"encoding/json"
	"net/http"

	log "github.com/sirupsen/logrus"
)

func HttpResponse(w http.ResponseWriter, code int, v interface{}) {
	w.WriteHeader(code)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Error(err)
		http.Error(w, err.Error(), code)
		return
	}
}

