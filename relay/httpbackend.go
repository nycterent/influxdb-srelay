package relay

import (
	"fmt"
	"time"

	"github.com/rs/zerolog"

	"github.com/toni-moreno/influxdb-srelay/config"
)

type dbBackend struct {
	cfg       *config.InfluxDBBackend
	clusterid string
	poster
	log       *zerolog.Logger
	inputType config.Input
	admin     string
}

func (b *dbBackend) getRetryBuffer() *retryBuffer {
	if p, ok := b.poster.(*retryBuffer); ok {
		return p
	}
	return nil
}

func NewDBBackend(cfg *config.InfluxDBBackend, l *zerolog.Logger, clustername string) (*dbBackend, error) {

	ret := &dbBackend{cfg: cfg, log: l}

	// Set a timeout
	timeout := DefaultHTTPTimeout
	if cfg.Timeout != "" {
		t, err := time.ParseDuration(cfg.Timeout)
		if err != nil {
			return nil, fmt.Errorf("error parsing HTTP timeout '%v'", err)
		}
		timeout = t
	}

	// Get underlying Poster instance
	var p poster = newSimplePoster(cfg.Name, cfg.Location, clustername, timeout, cfg.SkipTLSVerification)

	// If configured, create a retryBuffer per backend.
	// This way we serialize retries against each backend.
	if cfg.BufferSizeMB > 0 {
		max := DefaultMaxDelayInterval
		if cfg.MaxDelayInterval != "" {
			m, err := time.ParseDuration(cfg.MaxDelayInterval)
			if err != nil {
				return nil, fmt.Errorf("error parsing max retry time %v", err)
			}
			max = m
		}

		batch := DefaultBatchSizeKB * KB
		if cfg.MaxBatchKB > 0 {
			batch = cfg.MaxBatchKB * KB
		}

		p = newRetryBuffer(cfg.BufferSizeMB*MB, batch, max, p)
	}

	/*return &dbBackend{
		poster: p,
	}, nil*/
	ret.poster = p
	return ret, nil
}
