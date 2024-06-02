package service

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	cmtbytes "github.com/cometbft/cometbft/libs/bytes"
	cmtpubsub "github.com/cometbft/cometbft/libs/pubsub"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"

	"github.com/cometbft/cometbft/crypto"
	"github.com/cometbft/cometbft/types"
	"github.com/rangesecurity/ctop/cred"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
)

type RedisEventStream struct {
	CredClient *cred.CredClient

	ctx    context.Context
	cancel context.CancelFunc
}

func NewRedisEventStream(
	ctx context.Context,
	redisUrl string,
	unsafe bool,
) (*RedisEventStream, error) {
	ctx, cancel := context.WithCancel(ctx)
	cc, err := cred.New(ctx, redisUrl, unsafe)
	if err != nil {
		cancel()
		return nil, err
	}
	return &RedisEventStream{
		CredClient: cc,
		ctx:        ctx,
		cancel:     cancel,
	}, nil
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
				log.Err(err).Msg("failed to read redis stream")
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
								fmt.Printf("vote parse failed %+v\n", err)
								log.Error().Err(err).Msg("failed to parse redis value to vote object")
								continue
							}
							outCh <- voteInfo
						} else if isNewRound {
							newRound, err := parseRedisValueToNewRound(blockHeight, message.Values)
							if err != nil {
								fmt.Printf("new round parse failed %+v\n", err)
								log.Error().Err(err).Msg("failed to parse redis value to new round object")
							}
							outCh <- newRound
						} else if isNewRoundStep {
							roundState, err := parseRedisValueToRoundState(blockHeight, message.Values)
							if err != nil {
								fmt.Printf("round state parse failed %+v\n", err)
								log.Error().Err(err).Msg("failed to parse redis value to round state object")
							}
							outCh <- roundState
						} else {
							log.Error().Msg("event type is neither vote, new_round, or new_round_step. this is unexpected")
						}

					}
				}
			}

		}

	}()
	return nil
}

func parseRedisValueToNewRound(blockHeight int64, values map[string]interface{}) (*types.EventDataNewRound, error) {
	var (
		err           error
		ok            bool
		round         int64
		step          string
		proposer      string
		proposerIndex int64
	)
	if values["round"] == nil {
		return nil, fmt.Errorf("round is nil")
	} else if round_, ok := values["round"].(string); !ok {
		return nil, fmt.Errorf("failed to parse round")
	} else {
		round, err = strconv.ParseInt(round_, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("round ParseInt failed %s", err)
		}
	}

	if values["step"] == nil {
		return nil, fmt.Errorf("step is nil")
	} else if step, ok = values["step"].(string); !ok {
		return nil, fmt.Errorf("failed to parse step")
	}

	if values["proposer"] == nil {
		return nil, fmt.Errorf("proposer is nil")
	} else if proposer, ok = values["proposer"].(string); !ok {
		return nil, fmt.Errorf("failed to parse proposer")
	}

	if values["proposer_index"] == nil {
		return nil, fmt.Errorf("proposer_index is nil")
	} else if proposerIndex_, ok := values["proposer_index"].(string); !ok {
		return nil, fmt.Errorf("failed to parse proposer_index")
	} else {
		proposerIndex, err = strconv.ParseInt(proposerIndex_, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("proposer_index ParseInt failed")
		}
	}

	return &types.EventDataNewRound{
		Height: blockHeight,
		Round:  int32(round),
		Step:   step,
		Proposer: types.ValidatorInfo{
			Address: crypto.AddressHash([]byte(proposer)),
			Index:   int32(proposerIndex),
		},
	}, nil

}

func parseRedisValueToRoundState(blockHeight int64, values map[string]interface{}) (*types.EventDataRoundState, error) {
	var (
		err   error
		ok    bool
		round int64
		step  string
	)

	if values["round"] == nil {
		return nil, fmt.Errorf("round is nil")
	} else if round_, ok := values["round"].(string); !ok {
		return nil, fmt.Errorf("failed to parse round")
	} else {
		round, err = strconv.ParseInt(round_, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("round ParseInt failed %s", err)
		}
	}

	if values["step"] == nil {
		return nil, fmt.Errorf("step is nil")
	} else if step, ok = values["step"].(string); !ok {
		return nil, fmt.Errorf("failed to parse step")
	}
	return &types.EventDataRoundState{
		Height: blockHeight,
		Round:  int32(round),
		Step:   step,
	}, nil
}

func parseRedisValueToVote(blockHeight int64, values map[string]interface{}) (*types.Vote, error) {
	var (
		err              error
		ok               bool
		type_            string
		round            int64
		blockID          string
		timestamp        time.Time
		validatorAddress types.Address
		validatorIndex   int64
		signature        []byte
	)
	if values["type"] == nil {
		return nil, fmt.Errorf("type is nil")
	} else if type_, ok = values["type"].(string); !ok {
		return nil, fmt.Errorf("failed to parse type")
	}
	if values["round"] == nil {
		return nil, fmt.Errorf("round is nil")
	} else if round_, ok := values["round"].(string); !ok {
		return nil, fmt.Errorf("failed to parse round")
	} else {
		round, err = strconv.ParseInt(round_, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("round ParseInt failed %s", err)
		}

	}
	if values["block_hash"] == nil {
		return nil, fmt.Errorf("block_hash is nil")
	} else if blockID, ok = values["block_hash"].(string); !ok {
		return nil, fmt.Errorf("failed to parse block_hash")
	}
	if values["timestamp"] == nil {
		return nil, fmt.Errorf("timestamp is nil")
	} else if timestamp_, ok := values["timestamp"].(string); !ok {
		return nil, fmt.Errorf("failed to parse timestamp")
	} else {
		timestamp, err = time.Parse("2006-01-02 15:04:05.999999999 -0700 MST", timestamp_)
		if err != nil {
			return nil, fmt.Errorf("failed to parse time %v", err)
		}
	}

	if values["validator"] == nil {
		return nil, fmt.Errorf("validator is nil")
	} else if validatorAddress_, ok := values["validator"].(string); !ok {
		return nil, fmt.Errorf("failed to parse validator")
	} else {
		validatorAddress = crypto.AddressHash([]byte(validatorAddress_))
	}

	if values["index"] == nil {
		return nil, fmt.Errorf("validate index is nil")
	} else if index_, ok := values["index"].(string); !ok {
		return nil, fmt.Errorf("failed to parse index")
	} else {
		validatorIndex, err = strconv.ParseInt(index_, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("failed to parse index %v", err)
		}
	}

	if values["signature"] == nil {
		return nil, fmt.Errorf("signature is nil")
	} else if signature_, ok := values["signature"].(string); !ok {
		return nil, fmt.Errorf("failed to parse signature")
	} else {
		signature = []byte(signature_)
	}

	blockParts := strings.Split(blockID, ":")
	if len(blockParts) != 3 {
		return nil, fmt.Errorf("failed to parse blockId")
	}
	blockIdHash := cmtbytes.HexBytes{}
	if err := blockIdHash.Unmarshal([]byte(blockParts[0])); err != nil {
		return nil, err
	}
	partTotal, err := strconv.ParseUint(blockParts[1], 10, 32)
	if err != nil {
		return nil, err
	}
	partHash := cmtbytes.HexBytes{}
	if err := blockIdHash.Unmarshal([]byte(blockParts[2])); err != nil {
		return nil, err
	}
	return &types.Vote{
		Type:   cmtproto.SignedMsgType(cmtproto.SignedMsgType_value[type_]),
		Round:  int32(round),
		Height: blockHeight,
		BlockID: types.BlockID{
			Hash: blockIdHash,
			PartSetHeader: types.PartSetHeader{
				Total: uint32(partTotal),
				Hash:  partHash,
			},
		},
		Timestamp:        timestamp,
		ValidatorAddress: validatorAddress,
		ValidatorIndex:   int32(validatorIndex),
		Signature:        signature,
	}, nil
}
