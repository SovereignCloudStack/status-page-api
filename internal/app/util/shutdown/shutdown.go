package shutdown

import (
	"context"
	"sync"
	"time"

	metricsServer "github.com/SovereignCloudStack/status-page-api/internal/app/metrics"
	apiServer "github.com/SovereignCloudStack/status-page-api/internal/app/server"
	"github.com/rs/zerolog"
)

// Shutdown gracefully shutdowns all services in the timeout duration.
func Shutdown(
	timeout time.Duration,
	apiServer *apiServer.Server,
	metricsServer *metricsServer.Server,
	logger *zerolog.Logger,
) {
	var waitGroup sync.WaitGroup

	ctx, cancel := context.WithTimeout(context.Background(), timeout)

	numberOfServices := 2
	waitGroup.Add(numberOfServices)

	go func() {
		defer waitGroup.Done()

		err := metricsServer.Shutdown(ctx)
		if err != nil {
			logger.Warn().Err(err).Msg("error shutting down metrics server")
		}
	}()

	go func() {
		defer waitGroup.Done()

		err := apiServer.Shutdown(ctx)
		if err != nil {
			logger.Warn().Err(err).Msg("error shutting down server")
		}
	}()

	waitGroup.Wait()
	cancel()
}
