package cred

import (
	"context"
	"fmt"

	"github.com/cometbft/cometbft/types"
	"github.com/redis/go-redis/v9"
)

type CredClient struct {
	unsafe bool
	rdb    *redis.Client
}

// create a new CredClient, if unsafe is true allows running the flushdb command
func New(ctx context.Context, url string, unsafe bool) (*CredClient, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr: url,
	})
	scripts := [3]*redis.Script{VotesLuaScript, NewRoundStepScript, NewRoundScript}
	for _, script := range scripts {
		res := script.Load(ctx, rdb)
		if err := res.Err(); err != nil {
			return nil, fmt.Errorf("failed to load script %s", err)
		}
	}
	return &CredClient{
		unsafe,
		rdb,
	}, nil
}

func (c *CredClient) StoreVote(ctx context.Context, network string, voteInfo types.EventDataVote) error {
	return VotesLuaScript.Run(
		ctx,
		c.rdb,
		nil,
		[]interface{}{
			network,
			voteInfo.Vote.Height,
			voteInfo.Vote.ValidatorAddress.String(),
			voteInfo.Vote.Round,
			voteInfo.Vote.Signature,
			voteInfo.Vote.ValidatorIndex,
			voteInfo.Vote.BlockID.String(),
			voteInfo.Vote.Type.String(),
			voteInfo.Vote.Timestamp.String(),
		},
	).Err()
}

func (c *CredClient) StoreNewRound(ctx context.Context, network string, roundInfo types.EventDataNewRound) error {

	return NewRoundScript.Run(
		ctx,
		c.rdb,
		nil,
		[]interface{}{
			network,
			roundInfo.Height,
			roundInfo.Round,
			roundInfo.Step,
			roundInfo.Proposer.Address.String(),
			roundInfo.Proposer.Index,
		},
	).Err()
}

func (c *CredClient) StoreNewRoundStep(ctx context.Context, network string, roundInfo types.EventDataRoundState) error {

	return NewRoundStepScript.Run(
		ctx,
		c.rdb,
		nil,
		[]interface{}{
			network,
			roundInfo.Height,
			roundInfo.Round,
			roundInfo.Step,
		},
	).Err()
}
func (c *CredClient) FlushAll(ctx context.Context) error {
	if c.unsafe {
		return c.rdb.FlushAll(ctx).Err()
	} else {
		return nil
	}
}

func (c *CredClient) Redis() *redis.Client { return c.rdb }
