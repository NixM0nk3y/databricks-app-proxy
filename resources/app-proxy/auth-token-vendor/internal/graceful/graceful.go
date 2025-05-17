package graceful

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"
	"tokenvendor/internal/api"

	"github.com/rs/zerolog/log"
)

type graceful struct {
	Context context.Context
	Done    context.CancelFunc
	Api     *api.API
}

func (gf *graceful) SignalHandle() error {
	signalChannel := gf.getStopSignalsChannel()
	select {
	case sig := <-signalChannel:
		log.Warn().Msgf("received signal: %s", sig)
		gf.Shutdown()
		gf.Done()
	case <-gf.Context.Done():
		log.Warn().Msg("closing signal goroutine")
		gf.Shutdown()
		return gf.Context.Err()
	}
	return nil
}

func (gf *graceful) Shutdown() {
	log.Warn().Msgf("starting api shutdown")
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(gf.Api.Config.Server.ShutTime)*time.Second)
	defer cancel()
	gf.Api.Server.Shutdown(ctx)
}

func (gf *graceful) getStopSignalsChannel() <-chan os.Signal {
	signalChannel := make(chan os.Signal, 1)
	signal.Notify(signalChannel,
		os.Interrupt,
		syscall.SIGQUIT,
		syscall.SIGHUP,
		syscall.SIGTERM,
	)
	return signalChannel
}

func NewGraceful(api *api.API) *graceful {

	// ErrGroup for graceful shutdown
	ctx, done := context.WithCancel(context.Background())

	eg := &graceful{
		Context: ctx,
		Done:    done,
		Api:     api,
	}

	return eg
}
