package routes

import (
	"net/http"

	"tokenvendor/pkg/version"

	"github.com/go-chi/render"
	"github.com/rs/zerolog/log"
)

func VersionHandler(w http.ResponseWriter, r *http.Request) {

	logger := log.Ctx(r.Context())

	logger.Info().Msg("version endpoint called")

	render.Status(r, http.StatusOK)

	render.Render(w, r, &version.VersionResponse{
		Version:   version.Version,
		BuildHash: version.BuildHash,
		BuildDate: version.BuildDate,
	})
}
