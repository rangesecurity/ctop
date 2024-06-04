package service_test

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/joho/godotenv"
	"github.com/rangesecurity/ctop/cmd/bun/migrations"
	"github.com/rangesecurity/ctop/db"
	"github.com/rangesecurity/ctop/service"
	"github.com/stretchr/testify/require"
	"github.com/uptrace/bun/migrate"
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
			_, err := database.DB.NewDropTable().Model(model).Exec(context.Background())
			if err != nil {
				fmt.Println("failed to delete ", err)
			}
		}
	}
	recreate := func() {
		cleanUp()
		migrator := migrate.NewMigrator(database.DB, migrations.Migrations)
		migrator.Init(context.Background())
		database.CreateSchema(context.Background())
	}
	recreate()

	steps, err := database.GetNewRoundSteps(context.Background(), "osmosis")
	require.NoError(t, err)
	require.Equal(t, len(steps), 0)

	rounds, err := database.GetNewRounds(context.Background(), "osmosis")
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

	time.Sleep(time.Second * 20)

	votes, err := rds.Database.GetVotes(context.Background(), "osmosis")
	require.NoError(t, err)
	require.Greater(t, len(votes), 0)

	steps, err = rds.Database.GetNewRoundSteps(context.Background(), "osmosis")
	require.NoError(t, err)
	require.Greater(t, len(steps), 0)

	rounds, err = rds.Database.GetNewRounds(context.Background(), "osmosis")
	require.NoError(t, err)
	require.Greater(t, len(rounds), 0)

	cancel()

	wg.Wait()

	cleanUp()
}
