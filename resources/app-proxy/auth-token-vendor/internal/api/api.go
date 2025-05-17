package api

import (
	"context"
	"net/http"
	"time"

	"tokenvendor/internal/awstraceid"
	"tokenvendor/internal/caching"
	"tokenvendor/internal/chilogger"
	"tokenvendor/internal/config"
	"tokenvendor/internal/routes"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/go-chi/telemetry"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type API struct {
	Router *chi.Mux
	Server *http.Server
	Config *config.Config
}

var appCache = caching.NewAppCache()

func NewAPI(cfg *config.Config, logger zerolog.Logger) *API {

	router := chi.NewRouter()

	// use our custom logger
	router.Use(chilogger.LoggerMiddleware(
		chilogger.ChiLogger{
			Logger: &logger,
			Options: chilogger.ChiLoggerOptions{
				QuietRoutes: []string{"/health", "/metrics"},
				Tags: map[string]interface{}{
					"foo": "bar",
				},
			},
		},
	))

	// our AWS Request ID
	router.Use(awstraceid.AWStraceIdMiddleware)

	// recover from errors
	router.Use(middleware.Recoverer)

	// pass our config and cache via context
	router.Use(
		func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				ctx := context.WithValue(r.Context(), "appcache", appCache)
				r = r.WithContext(context.WithValue(ctx, config.ConfigKey, cfg))
				next.ServeHTTP(w, r)
			})
		},
	)

	// setup telemetry
	router.Use(telemetry.Collector(telemetry.Config{
		AllowAny: true,
	}, []string{"/token"}))

	// default to json
	router.Use(render.SetContentType(render.ContentTypeJSON))

	api := &API{
		Router: router,
		Server: &http.Server{
			Addr:         cfg.Server.Host,
			Handler:      router,
			ReadTimeout:  time.Duration(cfg.Server.ReadTime) * time.Second,
			WriteTimeout: time.Duration(cfg.Server.WriteTime) * time.Second,
		},
		Config: cfg,
	}

	api.setupRoutes()

	return api
}

func (api *API) setupRoutes() {
	api.Router.Get("/version", routes.VersionHandler)
	api.Router.Get("/health", routes.HealthHandler)
	api.Router.Get("/token", routes.TokenHandler)

	/*api.Router.Get("/panic", func(w http.ResponseWriter, r *http.Request) {
		panic("oh no")
	})*/

}

func (api *API) Run() error {
	log.Info().Msgf("starting API: %s", api.Server.Addr)
	return api.Server.ListenAndServe()
}
