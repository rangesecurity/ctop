package service

import (
	"context"
	"fmt"
	"sync"

	"github.com/rangesecurity/ctop/cred"
	"github.com/rs/zerolog/log"
)

type Service struct {
	CredClient *cred.CredClient
	connectors []*Connector
	ctx        context.Context
	cancel     context.CancelFunc

	sync.RWMutex
}

func NewService(
	ctx context.Context,
	redisUrl string,
	unsafe bool,
	// network_name -> rpc_url
	endpoints map[string]string,
) (*Service, error) {
	ctx, cancel := context.WithCancel(ctx)
	cc, err := cred.New(ctx, redisUrl, unsafe)
	if err != nil {
		cancel()
		return nil, err
	}
	connectors := make([]*Connector, 0, len(endpoints))

	for network, url := range endpoints {
		connector, err := NewConnector(ctx, network, url)
		if err != nil {
			cancel()
			return nil, err
		}
		connectors = append(connectors, connector)
	}

	return &Service{CredClient: cc, connectors: connectors, ctx: ctx, cancel: cancel}, nil
}

// starts all connectors and listens to incoming events
func (s *Service) StartEventSubscriptions() error {
	s.Lock()
	defer s.Unlock()
	for _, connector := range s.connectors {
		if err := connector.Start(); err != nil {
			return fmt.Errorf("failed to start connector %s", connector.network)
		}
		go func(connector *Connector) {
			network := connector.Network()
			for {
				select {
				case <-s.ctx.Done():
					return
				case voteInfo := <-connector.GetVotes():
					if err := s.CredClient.StoreVote(
						s.ctx,
						network,
						voteInfo,
					); err != nil {
						log.Error().Err(err).Msg("failed to store vote")
					}
				case roundInfo := <-connector.GetNewRounds():
					if err := s.CredClient.StoreNewRound(
						s.ctx,
						network,
						roundInfo,
					); err != nil {
						log.Error().Err(err).Msg("failed to store round")
					}
				case roundStep := <-connector.GetNewRoundSteps():
					if err := s.CredClient.StoreNewRoundStep(
						s.ctx,
						network,
						roundStep,
					); err != nil {
						log.Error().Err(err).Msg("failed to store round step")
					}
				}
			}
		}(connector)
	}
	return nil
}

func (s *Service) Close() {
	s.cancel()
}
