package service_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/go-pg/pg/v10/orm"
	"github.com/joho/godotenv"
	"github.com/rangesecurity/ctop/db"
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

	database, err := db.New("postgres://postgres:password123@localhost:5432/ctop")
	require.NoError(t, err)

	cleanUp := func() {
		models := []interface{}{
			(*db.VoteEvent)(nil),
			(*db.NewRoundEvent)(nil),
			(*db.NewRoundStepEvent)(nil),
		}
		for _, model := range models {
			database.DB.Model(model).DropTable(&orm.DropTableOptions{
				IfExists: true,
			})
		}
	}

	votes, err := database.GetVotes("osmosis")
	require.NoError(t, err)
	require.Equal(t, len(votes), 0)

	steps, err := database.GetNewRoundSteps("osmosis")
	require.NoError(t, err)
	require.Equal(t, len(steps), 0)

	rounds, err := database.GetNewRounds("osmosis")
	require.NoError(t, err)
	require.Equal(t, len(rounds), 0)

	rds, err := service.NewRedisEventStream(ctx, "localhost:6379", true, database)
	require.NoError(t, err)
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		require.NoError(t, rds.PersistNewRoundEvents("osmosis"))
	}()
	wg.Add(1)
	go func() {
		defer wg.Done()
		require.NoError(t, rds.PersistNewRoundStepEvents("osmosis"))
	}()
	wg.Add(1)
	go func() {
		defer wg.Done()
		require.NoError(t, rds.PersistVoteEvents("osmosis"))
	}()

	time.Sleep(time.Second * 15)

	votes, err = rds.Database.GetVotes("osmosis")
	require.NoError(t, err)
	require.Greater(t, len(votes), 0)

	steps, err = rds.Database.GetNewRoundSteps("osmosis")
	require.NoError(t, err)
	require.Greater(t, len(steps), 0)

	rounds, err = rds.Database.GetNewRounds("osmosis")
	require.NoError(t, err)
	require.Greater(t, len(rounds), 0)

	cancel()

	wg.Wait()

	cleanUp()
}
