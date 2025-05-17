package awstraceid

import (
	"net/http"

	"github.com/aws/aws-xray-sdk-go/header"
	"github.com/aws/aws-xray-sdk-go/xray"

	"github.com/rs/zerolog"
)

// Key to use when setting the request ID.
type ctxAWStraceId int

// AWStraceIdKey is the key that holds the unique request ID in a request context.
const AWStraceIdKey ctxAWStraceId = 0

func AWStraceIdMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		awsTraceHeader := r.Header.Get("X-Amzn-Trace-Id")
		var id string
		if awsTraceHeader != "" {
			traceData := header.FromString(awsTraceHeader)
			id = traceData.TraceID
		} else {
			id = xray.NewTraceID()
		}

		// fetch the logger from context and update the context
		// with the correlation id value
		logger := zerolog.Ctx(ctx)
		logger.UpdateContext(func(c zerolog.Context) zerolog.Context {
			return c.Str("correlation_id", id)
		})

		// set the response header
		w.Header().Set("X-Correlation-Id", id)
		next.ServeHTTP(w, r)
	})
}
