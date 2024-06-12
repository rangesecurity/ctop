package cred_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"

	"github.com/cometbft/cometbft/crypto"
	"github.com/cometbft/cometbft/crypto/tmhash"
	"github.com/cometbft/cometbft/types"
	"github.com/rangesecurity/ctop/cred"
	"github.com/stretchr/testify/require"
)

func TestCred(t *testing.T) {
	// TODO: use table tests

	ctx := context.Background()
	client, err := cred.New(ctx, "localhost:6379", true)
	require.NoError(t, err)

	// clear pre-existing state
	require.NoError(t, client.FlushAll(ctx))

	// test votes
	startHeight := int64(12345)
	for i := startHeight; i < 12348; i++ {
		require.NoError(t, client.StoreVote(ctx, "osmosis", types.EventDataVote{
			Vote: exampleVote(i, byte(cmtproto.PrevoteType)),
		}))
		require.NoError(t, client.StoreVote(ctx, "osmosis", types.EventDataVote{
			Vote: exampleVote(i, byte(cmtproto.PrevoteType)),
		}))
		msgs, err := client.Redis().XRange(
			ctx, "osmosis:votes", fmt.Sprint(i), fmt.Sprint(i),
		).Result()
		require.Len(t, msgs, 2)
		require.NoError(t, err)
		// starts at 1, since the lua script calls INCR
		require.Equal(t, msgs[0].ID, fmt.Sprintf("%v-1", i))
		require.Equal(t, msgs[1].ID, fmt.Sprintf("%v-2", i))
	}
	mockKey1 := types.NewMockPV()
	mockKey2 := types.NewMockPV()
	validator1 := types.NewValidator(mockKey1.PrivKey.PubKey(), 10)
	validator2 := types.NewValidator(mockKey2.PrivKey.PubKey(), 11)
	// test round
	require.NoError(t, client.StoreNewRound(ctx, "osmosis", types.EventDataNewRound{
		Height: 112345,
		Round:  0,
		Step:   "step",
		Proposer: types.ValidatorInfo{
			Address: validator1.Address,
			Index:   1,
		},
	}))
	require.NoError(t, client.StoreNewRound(ctx, "osmosis", types.EventDataNewRound{
		Height: 112345,
		Round:  0,
		Step:   "step",
		Proposer: types.ValidatorInfo{
			Address: validator2.Address,
			Index:   2,
		},
	}))
	msgs, err := client.Redis().XRange(
		ctx, "osmosis:new_round", fmt.Sprint(112345), fmt.Sprint(112345),
	).Result()
	require.NoError(t, err)
	require.Len(t, msgs, 2)
	require.Equal(t, msgs[0].ID, "112345-1")
	require.Equal(t, msgs[1].ID, "112345-2")

	require.NoError(t, client.StoreNewRoundStep(
		ctx,
		"osmosis",
		types.EventDataRoundState{
			Height: 22345,
			Round:  1,
			Step:   "RoundStepPropose",
		},
	))
	require.NoError(t, client.StoreNewRoundStep(
		ctx,
		"osmosis",
		types.EventDataRoundState{
			Height: 22345,
			Round:  1,
			Step:   "RoundStepPrecommit",
		},
	))
	msgs, err = client.Redis().XRange(
		ctx, "osmosis:new_round_step", fmt.Sprint(22345), fmt.Sprint(22345),
	).Result()
	require.NoError(t, err)
	require.Len(t, msgs, 2)
	require.Equal(t, msgs[0].ID, "22345-1")
	require.Equal(t, msgs[1].ID, "22345-2")
}

func exampleVote(height int64, t byte) *types.Vote {
	stamp, err := time.Parse(types.TimeFormat, "2017-12-25T03:00:01.234Z")
	if err != nil {
		panic(err)
	}

	return &types.Vote{
		Type:      cmtproto.SignedMsgType(t),
		Height:    height,
		Round:     2,
		Timestamp: stamp,
		BlockID: types.BlockID{
			Hash: tmhash.Sum([]byte("blockID_hash")),
			PartSetHeader: types.PartSetHeader{
				Total: 1000000,
				Hash:  tmhash.Sum([]byte("blockID_part_set_header_hash")),
			},
		},
		ValidatorAddress: crypto.AddressHash([]byte("validator_address")),
		ValidatorIndex:   56789,
	}
}
