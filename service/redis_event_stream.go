package service

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	cmtpubsub "github.com/cometbft/cometbft/libs/pubsub"
	"github.com/cometbft/cometbft/types"
	"github.com/rangesecurity/ctop/common"
	"github.com/rangesecurity/ctop/cred"
	"github.com/rangesecurity/ctop/db"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
)

type RedisEventStream struct {
	CredClient *cred.CredClient
	Database   *db.Database
	ctx        context.Context
	cancel     context.CancelFunc
}

func NewRedisEventStream(
	ctx context.Context,
	redisUrl string,
	unsafe bool,
	database *db.Database,
) (*RedisEventStream, error) {
	ctx, cancel := context.WithCancel(ctx)
	cc, err := cred.New(ctx, redisUrl, unsafe)
	if err != nil {
		cancel()
		return nil, err
	}
	return &RedisEventStream{
		CredClient: cc,
		Database:   database,
		ctx:        ctx,
		cancel:     cancel,
	}, nil
}

func (rds *RedisEventStream) PersistVoteEvents(
	network string,
) error {
	outCh := make(chan interface{}, 1024)
	if err := rds.StreamRedisEvents(
		network,
		types.EventQueryVote,
		outCh,
	); err != nil {
		return fmt.Errorf("failed to stream redis events %+v", err)
	}
	for {
		select {
		case msg := <-outCh:
			voteInfo, ok := msg.(*common.ParsedVote)
			if !ok {
				log.Error().Msg("unexpected msg type")
				continue
			}
			if err := rds.Database.StoreVote(
				rds.ctx,
				network,
				*voteInfo,
			); err != nil {
				log.Error().Err(err).Msg("failed to store vote")
			}
		case <-rds.ctx.Done():
			return nil
		}
	}
}

func (rds *RedisEventStream) PersistNewRoundEvents(
	network string,
) error {
	outCh := make(chan interface{}, 256)
	if err := rds.StreamRedisEvents(
		network,
		types.EventQueryNewRound,
		outCh,
	); err != nil {
		return fmt.Errorf("failed to stream redis events %+v", err)
	}
	for {
		select {
		case msg := <-outCh:
			roundInfo, ok := msg.(*common.ParsedNewRound)
			if !ok {
				log.Error().Msg("unexpected msg type")
				continue
			}
			if err := rds.Database.StoreNewRound(
				rds.ctx,
				network,
				*roundInfo,
			); err != nil {
				log.Error().Err(err).Msg("failed to store round")
			}
		case <-rds.ctx.Done():
			return nil
		}
	}
}

func (rds *RedisEventStream) PersistNewRoundStepEvents(
	network string,
) error {
	outCh := make(chan interface{}, 256)
	if err := rds.StreamRedisEvents(
		network,
		types.EventQueryNewRoundStep,
		outCh,
	); err != nil {
		return fmt.Errorf("failed to stream redis events %+v", err)
	}
	for {
		select {
		case msg := <-outCh:
			roundState, ok := msg.(*common.ParsedNewRoundStep)
			if !ok {
				log.Error().Msg("unexpected msg type")
				continue
			}
			if err := rds.Database.StoreNewRoundStep(
				rds.ctx,
				network,
				*roundState,
			); err != nil {
				log.Error().Err(err).Msg("failed to store round state")
			}
		case <-rds.ctx.Done():
			return nil
		}
	}
}

// streams events from redis, inserting into database
func (rds *RedisEventStream) StreamRedisEvents(
	network string,
	eventType cmtpubsub.Query,
	outCh chan interface{},
) error {
	var (
		streamKey      string
		isVote         bool
		isNewRound     bool
		isNewRoundStep bool
	)
	if ok, _ := eventType.Matches(map[string][]string{"tm.event": {"Vote"}}); ok {
		streamKey = ":votes"
		isVote = true

	} else if ok, _ := eventType.Matches(map[string][]string{"tm.event": {"NewRound"}}); ok {
		streamKey = ":new_round"
		isNewRound = true
	} else if ok, _ := eventType.Matches(map[string][]string{"tm.event": {"NewRoundStep"}}); ok {
		streamKey = ":new_round_step"
		isNewRoundStep = true
	} else {
		return fmt.Errorf("unsupported event %s", eventType)
	}

	go func() {
		for {
			entries, err := rds.CredClient.Redis().XRead(
				rds.ctx,
				&redis.XReadArgs{
					Streams: []string{network + streamKey, "0"},
					Block:   0,
				},
			).Result()
			if err != nil {
				log.Err(err).Str("event.type", streamKey[1:]).Msg("failed to read redis stream")
				continue
			}
			select {
			case <-rds.ctx.Done():
				return
			default:
				for _, stream := range entries {
					for _, message := range stream.Messages {
						parts := strings.Split(message.ID, "-")
						if len(parts) != 2 {
							// improperly formatted id
							continue
						}
						blockHeight, err := strconv.ParseInt(parts[0], 10, 64)
						if err != nil {
							continue
						}
						if isVote {
							voteInfo, err := parseRedisValueToVote(blockHeight, message.Values)
							if err != nil {
								log.Error().Err(err).Str("event.type", streamKey[1:]).Msg("failed to parse redis value to vote object")
								continue
							}
							outCh <- voteInfo
						} else if isNewRound {
							newRound, err := parseRedisValueToNewRound(blockHeight, message.Values)
							if err != nil {
								log.Error().Err(err).Str("event.type", streamKey[1:]).Msg("failed to parse redis value to new round object")
								continue
							}
							outCh <- newRound
						} else if isNewRoundStep {
							roundState, err := parseRedisValueToRoundState(blockHeight, message.Values)
							if err != nil {
								log.Error().Err(err).Str("event.type", streamKey[1:]).Msg("failed to parse redis value to round state object")
								continue
							}
							outCh <- roundState
						} else {
							log.Error().Str("event.type", streamKey).Msg("event type is neither vote, new_round, or new_round_step. this is unexpected")
							continue
						}
						// clear out messages which we pulled from redis
						if err := rds.CredClient.Redis().XDel(rds.ctx, network+streamKey, message.ID).Err(); err != nil {
							log.Error().Err(err).Str("event.type", streamKey[1:]).Str("id", message.ID).Msg("failed to clear message from stream")
						}
					}
				}
			}

		}

	}()
	return nil
}
