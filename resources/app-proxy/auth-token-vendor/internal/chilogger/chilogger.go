package chilogger

import (
	"net/http"
	"runtime/debug"
	"time"

	"tokenvendor/internal/utils"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type ChiLoggerOptions struct {

	// Tags are additional fields included at the root level of all logs.
	// These can be useful for example the commit hash of a build, or an environment
	// name like prod/stg/dev
	Tags map[string]interface{}

	// QuietDownRoutes are routes which are temporarily excluded from logging for a QuietDownPeriod after it occurs
	// for the first time
	// to cancel noise from logging for routes that are known to be noisy.
	QuietRoutes []string
}

type ChiLogger struct {
	Logger  *zerolog.Logger
	Options ChiLoggerOptions
}

func LoggerMiddleware(cl ChiLogger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {

			//
			l := log.With().Fields(cl.Options.Tags).Logger()

			// skip logging quiet routes
			if utils.InArray(cl.Options.QuietRoutes, r.URL.Path) {
				next.ServeHTTP(w, r)
				return
			}

			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

			start := time.Now()

			// wrap the request in our logging context
			r = r.WithContext(l.WithContext(r.Context()))

			defer func() {

				// update logger context
				logger := zerolog.Ctx(r.Context())

				// Recover and record stack traces in case of a panic
				if rec := recover(); rec != nil {
					logger.Error().
						Str("type", "error").
						Interface("recover_info", rec).
						Bytes("debug_stack", debug.Stack()).
						Msg("log system error")
					http.Error(ww, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				}

				// log end request
				logger.Info().
					Str("type", "access").
					Fields(map[string]interface{}{
						"remote_ip":  r.RemoteAddr,
						"url":        r.URL.Path,
						"proto":      r.Proto,
						"method":     r.Method,
						"user_agent": r.Header.Get("User-Agent"),
						"status":     ww.Status(),
						"latency_ms": float64(time.Since(start).Nanoseconds()) / 1000000.0,
						"bytes_in":   r.Header.Get("Content-Length"),
						"bytes_out":  ww.BytesWritten(),
					}).
					Msg("incoming_request")
			}()

			next.ServeHTTP(ww, r)
		}
		return http.HandlerFunc(fn)
	}
}
