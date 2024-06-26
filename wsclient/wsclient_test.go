package wsclient_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/cometbft/cometbft/types"
	"github.com/joho/godotenv"
	"github.com/rangesecurity/ctop/wsclient"
	"github.com/stretchr/testify/require"
)

func TestWsClientSubscribeVotes(t *testing.T) {
	envs, err := godotenv.Read("../.env")
	require.NoError(t, err)
	rpcUrl := envs["OSMOSIS_RPC"]
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	client, err := wsclient.NewClient(rpcUrl)
	require.NoError(t, err)
	outCh, err := client.SubscribeVotes(ctx)
	require.NoError(t, err)
	defer client.UnsubscribeVotes(ctx)
	i := 0
	fmt.Println("Listening for events")
	for {
		vote := <-outCh
		voteInfo, ok := vote.Data.(types.EventDataVote)
		if !ok {
			continue
		}
		t.Logf(
			"type %s, height %v, round %v, blockId %s, timestamp %s, validator %s, validatorIndex %v\n",
			voteInfo.Vote.Type, voteInfo.Vote.Height, voteInfo.Vote.Round, voteInfo.Vote.BlockID, voteInfo.Vote.Timestamp, voteInfo.Vote.ValidatorAddress, voteInfo.Vote.ValidatorIndex,
		)
		i++
		if i >= 3 {
			break
		}
	}
}

func TestWsClientSubscribeNewRound(t *testing.T) {
	envs, err := godotenv.Read("../.env")
	require.NoError(t, err)
	rpcUrl := envs["OSMOSIS_RPC"]
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	client, err := wsclient.NewClient(rpcUrl)
	require.NoError(t, err)
	outCh, err := client.SubscribeNewRound(ctx)
	require.NoError(t, err)
	defer client.UnsubscribeNewRound(ctx)
	i := 0
	fmt.Println("Listening for events")
	for {
		vote := <-outCh
		roundInfo, ok := vote.Data.(types.EventDataNewRound)
		if !ok {
			continue
		}
		t.Logf(
			"height %v, round %v, step %s, blockId %s\n",
			roundInfo.Height, roundInfo.Round, roundInfo.Step, roundInfo.Proposer.Address,
		)
		i++
		if i >= 3 {
			break
		}
	}
}

func TestWsClientSubscribeNewRoundStep(t *testing.T) {
	envs, err := godotenv.Read("../.env")
	require.NoError(t, err)
	rpcUrl := envs["OSMOSIS_RPC"]
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	client, err := wsclient.NewClient(rpcUrl)
	require.NoError(t, err)
	outCh, err := client.SubscribeNewRoundStep(ctx)
	require.NoError(t, err)
	defer client.UnsubscribeNewRoundStep(ctx)
	i := 0
	fmt.Println("Listening for events")
	for {
		vote := <-outCh
		roundInfo, ok := vote.Data.(types.EventDataRoundState)
		if !ok {
			continue
		}
		t.Logf(
			"height %v, round %v, step %s\n",
			roundInfo.Height, roundInfo.Round, roundInfo.Step,
		)
		i++
		if i >= 3 {
			break
		}
	}
}

func TestWsClientValidator(t *testing.T) {
	envs, err := godotenv.Read("../.env")
	require.NoError(t, err)
	rpcUrl := envs["OSMOSIS_RPC"]
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	client, err := wsclient.NewClient(rpcUrl)
	require.NoError(t, err)
	valis, err := client.Validators(ctx)
	require.NoError(t, err)
	require.Greater(t, len(valis), 0)
}
