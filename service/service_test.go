package service_test

import (
	"context"
	"testing"

	"github.com/cometbft/cometbft/types"
	"github.com/joho/godotenv"
	"github.com/rangesecurity/ctop/service"
	"github.com/stretchr/testify/require"
)

func TestService(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	envs, err := godotenv.Read("../.env")
	require.NoError(t, err)
	rpcUrl := envs["OSMOSIS_RPC"]
	s, err := service.NewService(
		ctx,
		"localhost:6379",
		true,
		map[string]string{
			"osmosis": rpcUrl,
		},
	)
	require.NoError(t, err)
	require.NoError(t, s.CredClient.FlushAll(ctx))
	err = s.StartEventSubscriptions()
	require.NoError(t, err)
	var (
		receivedVoteEvent         = make(chan bool, 1)
		receivedNewRoundEvent     = make(chan bool, 1)
		receivedNewRoundStepEvent = make(chan bool, 1)
	)
	go func() {
		outCh := make(chan interface{}, 1024)
		require.NoError(t, s.StreamRedisEvents("osmosis", types.EventQueryVote, outCh))
		for {
			select {
			case msg := <-outCh:
				msg, ok := msg.(*types.Vote)
				require.True(t, ok)
				receivedVoteEvent <- true
			case <-ctx.Done():
				return
			}
		}
	}()
	go func() {
		outCh := make(chan interface{}, 1024)
		require.NoError(t, s.StreamRedisEvents("osmosis", types.EventQueryNewRound, outCh))
		for {
			select {
			case msg := <-outCh:
				msg, ok := msg.(*types.EventDataNewRound)
				require.True(t, ok)
				receivedNewRoundEvent <- true
			case <-ctx.Done():
				return
			}
		}
	}()
	go func() {
		outCh := make(chan interface{}, 1024)
		require.NoError(t, s.StreamRedisEvents("osmosis", types.EventQueryNewRoundStep, outCh))
		for {
			select {
			case msg := <-outCh:
				msg, ok := msg.(*types.EventDataRoundState)
				require.True(t, ok)
				receivedNewRoundStepEvent <- true
			case <-ctx.Done():
				return
			}
		}
	}()

	ok := <-receivedNewRoundEvent
	require.True(t, ok)
	ok = <-receivedVoteEvent
	require.True(t, ok)
	ok = <-receivedNewRoundStepEvent
	require.True(t, ok)

}
