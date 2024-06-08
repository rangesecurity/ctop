package service

import (
	"context"
	"time"

	"github.com/rangesecurity/ctop/db"
	"github.com/rs/zerolog/log"
)

// Periodically indexes
type ValidatorIndexer struct {
	connectors []*Connector
	db         *db.Database
	ctx        context.Context
	cancel     context.CancelFunc
}

func NewValidatorIndexer(
	ctx context.Context,
	db *db.Database,
	endpoints map[string]string,
) (*ValidatorIndexer, error) {
	ctx, cancel := context.WithCancel(ctx)
	connectors := make([]*Connector, 0, len(endpoints))

	for network, url := range endpoints {
		connector, err := NewConnector(ctx, network, url)
		if err != nil {
			cancel()
			return nil, err
		}
		connectors = append(connectors, connector)
	}
	return &ValidatorIndexer{
		connectors,
		db,
		ctx,
		cancel,
	}, nil
}

func (vi *ValidatorIndexer) Start(
	pollFrequency time.Duration,
) {
	ticker := time.NewTicker(pollFrequency)
	for {
		select {
		case <-vi.ctx.Done():
			return
		case <-ticker.C:
			for _, connector := range vi.connectors {
				valis, err := connector.Validators()
				if err != nil {
					log.Error().Err(err).Str("network", connector.Network()).Msg("failed to fetch validators")
				}
				data := make(map[string]interface{})
				for _, vali := range valis {
					data[vali.Address.String()] = struct{}{}
				}
				if err := vi.db.StoreOrUpdateValidators(
					vi.ctx, connector.Network(), data,
				); err != nil {
					log.Error().Err(err).Str("network", connector.Network()).Msg("failed to store validator")
				}
			}
		}
	}
}
