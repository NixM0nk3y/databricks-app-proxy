package routes

import (
	"net/http"

	"github.com/rs/zerolog/log"
)

func HealthHandler(w http.ResponseWriter, r *http.Request) {

	log.Debug().Msg("health endpoint called")

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}
