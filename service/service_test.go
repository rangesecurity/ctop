package service_test

import (
	"context"
	"testing"
	"time"

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
	outCh := make(chan interface{}, 1024)
	require.NoError(t, s.StreamRedisEvents("osmosis", types.EventQueryVote, outCh))
	require.NoError(t, s.StreamRedisEvents("osmosis", types.EventQueryNewRound, outCh))
	require.NoError(t, s.StreamRedisEvents("osmosis", types.EventQueryNewRoundStep, outCh))

	time.Sleep(time.Second * 10)

	msgs, err := s.CredClient.Redis().XRange(
		ctx,
		"osmosis:votes",
		"-", "+",
	).Result()
	require.NoError(t, err)
	require.Greater(t, len(msgs), 0)

	msgs, err = s.CredClient.Redis().XRange(
		ctx,
		"osmosis:new_round",
		"-", "+",
	).Result()
	require.NoError(t, err)
	require.Greater(t, len(msgs), 0)

	msgs, err = s.CredClient.Redis().XRange(
		ctx,
		"osmosis:new_round_step",
		"-", "+",
	).Result()
	require.NoError(t, err)
	require.Greater(t, len(msgs), 0)
	require.Greater(t, len(outCh), 0)

}
