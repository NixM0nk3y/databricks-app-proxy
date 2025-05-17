package routes

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"
	"tokenvendor/internal/caching"
	"tokenvendor/internal/config"
	"tokenvendor/internal/metrics"

	"github.com/go-chi/render"
	"github.com/rs/zerolog"
)

type TokenResponse struct {
	AccessToken string `json:"access_token"`
	Scope       string `json:"scope"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
}

type CacheEntry struct {
	Token   *TokenResponse
	Expires time.Time
}

const CACHEKEY = "access-token"

// number of seconds prior to the token expiring to refresh
const REFRESH_WINDOW_SECS = 120

func TokenHandler(w http.ResponseWriter, r *http.Request) {

	ctx := r.Context()

	logger := zerolog.Ctx(r.Context())

	metrics := metrics.NewAppMetrics("token")

	logger.Debug().Msg("token endpoint called")

	appCache := ctx.Value("appcache").(*caching.AppCache)

	var token *TokenResponse
	var cacheHit = false

	value, ok := appCache.Read(CACHEKEY)

	now := time.Now()

	if ok {
		// do we have a currently cached token?
		if now.Before(value.(*CacheEntry).Expires) {
			metrics.RecordAppHit("cache-hit")
			// no need to refresh
			cacheHit = true
			token = value.(*CacheEntry).Token
		}
	}

	if !cacheHit {

		metrics.RecordAppHit("cache-miss")

		logger.Info().Msg("cache miss - generating token")

		var tokenErr error

		token, tokenErr = GenerateAccessToken(ctx)

		// validate we've been successful
		if tokenErr != nil {

			// return a error
			render.Status(r, http.StatusForbidden)
			render.Render(w, r, &ErrorResponse{
				Status:  http.StatusForbidden,
				Message: "unexpected error received, try again later",
			})
			return
		}

		// add new token to the cache
		appCache.Update(CACHEKEY, &CacheEntry{
			Token:   token,
			Expires: now.Add(time.Second * time.Duration(token.ExpiresIn)),
		}, time.Second*time.Duration(token.ExpiresIn))
	}

	// add a authorisation header
	w.Header().Set("Authorization", fmt.Sprintf("Bearer %s", token.AccessToken))

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func GenerateAccessToken(ctx context.Context) (*TokenResponse, error) {

	logger := zerolog.Ctx(ctx)

	metrics := metrics.NewAppMetrics("access_token_generate")

	// pull our config from context
	cfg, ok := ctx.Value(config.ConfigKey).(*config.Config)
	if !ok || cfg == nil {
		logger.Panic().Msg("unable to retrieve config contest")
	}

	client := &http.Client{
		Timeout: time.Duration(cfg.Databricks.ReadTime) * time.Second,
	}

	// build our request
	postReq, err := http.NewRequestWithContext(ctx, http.MethodPost, fmt.Sprintf("%s/oidc/oauth2/v2.0/token", cfg.Databricks.Hostname), bytes.NewBuffer([]byte("grant_type=client_credentials&scope=all-apis")))
	if err != nil {
		logger.Panic().Stack().Err(err).Msg("failed to build post request")
	}
	postReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	postReq.SetBasicAuth(cfg.Databricks.ClientId, cfg.Databricks.ClientSecret)

	//
	metrics.RecordAppHit("requests")

	// execute our request
	logger.Info().Str("url", cfg.Databricks.Hostname).Msg("requesting auth token")
	start := time.Now()
	resp, err := client.Do(postReq)
	finish := time.Now()

	if err != nil {
		logger.Panic().Stack().Err(err).Msg("failed to build post request")
	}

	defer resp.Body.Close()

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Panic().Stack().Err(err).Msg("error reading response body")
	}

	// validate we've been successful
	if resp.StatusCode != http.StatusOK {

		// log and record a metric
		logger.Error().Int("status", resp.StatusCode).Str("response", string(body)).Msg("request failed")
		metrics.RecordAppHit("errors")

		return nil, errors.New("bad status returned from api")
	}

	// record a metric
	metrics.RecordDuration("request", start, finish)

	metrics.RecordAppHit("token_generate")

	logger.Debug().Str("body", string(body)).Msg("response")

	var token TokenResponse

	err = json.Unmarshal([]byte(body), &token)
	if err != nil {
		logger.Error().Str("response", string(body)).Err(err).Msg("unable to unmarshal token response")
		metrics.RecordAppHit("errors")
	}
	return &token, nil
}
