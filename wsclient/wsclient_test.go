package wsclient_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/cometbft/cometbft/types"
	"github.com/rangesecurity/ctop/wsclient"
	"github.com/stretchr/testify/require"
)

func TestWsClientSubscribeVotes(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	client, err := wsclient.NewClient("tcp://..:12557")
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
		if i >= 10 {
			break
		}
	}
}

func TestWsClientSubscribeNewRound(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	client, err := wsclient.NewClient("tcp://..:12557")
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
		if i >= 10 {
			break
		}
	}
}

func TestWsClientSubscribeNewRoundStep(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	client, err := wsclient.NewClient("tcp://..:12557")
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
		if i >= 10 {
			break
		}
	}
}
