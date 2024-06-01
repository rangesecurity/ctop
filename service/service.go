package service

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/cometbft/cometbft/crypto"
	cmtbytes "github.com/cometbft/cometbft/libs/bytes"

	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"

	cmtpubsub "github.com/cometbft/cometbft/libs/pubsub"
	"github.com/cometbft/cometbft/types"
	"github.com/redis/go-redis/v9"

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

// streams events from redis, inserting into database
func (s *Service) StreamRedisEvents(
	network string,
	eventType cmtpubsub.Query,
	outCh chan interface{},
) error {
	s.Lock()
	defer s.Unlock()
	if ok, _ := eventType.Matches(map[string][]string{"tm.event": {"Vote"}}); ok {
		go func() {
			for {
				entries, err := s.CredClient.Redis().XRead(
					s.ctx,
					&redis.XReadArgs{
						Streams: []string{network + ":votes", "0"},
						Block:   0,
					},
				).Result()
				if err != nil {
					log.Err(err).Msg("failed to read redis stream")
					continue
				}
				select {
				case <-s.ctx.Done():
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

							voteInfo, err := parseRedisValueToVote(blockHeight, message.Values)
							if err != nil {
								fmt.Printf("failed to parse value %+v\n", err)
								continue
							}
							outCh <- voteInfo
						}
					}
				}

			}

		}()
	} else if ok, _ := eventType.Matches(map[string][]string{"tm.event": {"NewRound"}}); ok {
		fmt.Println("streaming new round")
		// TODO
	} else if ok, _ := eventType.Matches(map[string][]string{"tm.event": {"NewRoundStep"}}); ok {
		fmt.Println("streaming new rond step")
		// TODO
	} else {
		return fmt.Errorf("unsupported event %s", eventType)
	}
	return nil
}

func (s *Service) Close() {
	s.cancel()
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
