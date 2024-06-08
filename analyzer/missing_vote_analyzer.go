package analyzer

import (
	"context"
	"time"

	"github.com/rangesecurity/ctop/db"
	"github.com/rs/zerolog/log"
)

// MissingVoteAnalyzer provides alerts whenever validators fail to vote within a specified timeframe
type MissingVoteAnalyzer struct {
	db     *db.Database
	ctx    context.Context
	cancel context.CancelFunc
}

func NewMissingVoteAnalyzer(
	ctx context.Context,
	db *db.Database,
) *MissingVoteAnalyzer {
	ctx, cancel := context.WithCancel(ctx)
	return &MissingVoteAnalyzer{
		db,
		ctx,
		cancel,
	}
}

func (mva *MissingVoteAnalyzer) Start(
	network string,
	pollFrequency time.Duration,
) {
	ticker := time.NewTicker(pollFrequency)
	for {
		select {
		case <-mva.ctx.Done():
			return
		case <-ticker.C:
			valis, err := mva.db.GetValidators(mva.ctx, network)
			if err != nil {
				log.Error().Err(err).Str("network", network).Msg("failed to get validators")
				continue
			}
			foundVotes := make(map[string]struct{})
			votes, err := mva.db.GetLatestVotesForNetwork(mva.ctx, network)
			if err != nil {
				log.Error().Err(err).Str("network", network).Msg("failed to query db for votes")
				continue
			}
			for _, vote := range votes {
				foundVotes[vote.ValidatorAddress] = struct{}{}
			}
			for validatorAddress := range valis.Data {
				if _, exists := foundVotes[validatorAddress]; !exists {
					log.Warn().Str("validator", validatorAddress).Msg("missing vote")
				}
			}
			log.Info().Int("num.validators_voted", len(foundVotes)).Int("total_validators", len(valis.Data)).Msg("checked votes")
		}
	}
}

func (mva *MissingVoteAnalyzer) Stop() {
	mva.cancel()
}
