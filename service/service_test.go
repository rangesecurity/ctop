package service_test

import (
	"context"
	"testing"
	"time"

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
}
