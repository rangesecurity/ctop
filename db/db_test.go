package db_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/uptrace/bun/migrate"

	"github.com/cometbft/cometbft/crypto"
	"github.com/cometbft/cometbft/crypto/tmhash"
	"github.com/cometbft/cometbft/types"
	"github.com/rangesecurity/ctop/cmd/bun/migrations"
	"github.com/rangesecurity/ctop/db"
	"github.com/stretchr/testify/require"
)

func TestDb(t *testing.T) {
	url := "postgres://postgres:password123@localhost:5432/ctop"
	database, err := db.New(url)
	require.NoError(t, err)
	cleanUp := func() {
		models := []interface{}{
			(*db.VoteEvent)(nil),
			(*db.NewRoundEvent)(nil),
			(*db.NewRoundStepEvent)(nil),
			(*db.Validators)(nil),
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
		migrator.Reset(context.Background())
		database.CreateSchema(context.Background())
	}
	recreate()
	require.NoError(t, database.StoreVote(context.Background(), "osmosis", *exampleVote(12345, byte(cmtproto.PrevoteType))))
	votes, err := database.GetVotes(context.Background(), "osmosis")
	require.NoError(t, err)
	require.Len(t, votes, 1)

	mockKey1 := types.NewMockPV()
	validator1 := types.NewValidator(mockKey1.PrivKey.PubKey(), 10)
	require.NoError(t, database.StoreNewRound(context.Background(), "osmosis", types.EventDataNewRound{
		Height: 112345,
		Round:  0,
		Step:   "step",
		Proposer: types.ValidatorInfo{
			Address: validator1.Address,
			Index:   1,
		},
	}))

	newRounds, err := database.GetNewRounds(context.Background(), "osmosis")
	require.NoError(t, err)
	require.Len(t, newRounds, 1)

	require.NoError(t, database.StoreNewRoundStep(
		context.Background(),
		"osmosis",
		types.EventDataRoundState{
			Height: 11234,
			Round:  0,
			Step:   "RoundStepPropose",
		},
	))
	roundSteps, err := database.GetNewRoundSteps(context.Background(), "osmosis")
	require.NoError(t, err)
	require.Len(t, roundSteps, 1)

	data := map[string]interface{}{
		"validator1": time.Unix(0, 0),
		"validator2": time.Unix(0, 0),
	}

	require.NoError(t, database.StoreOrUpdateValidators(context.Background(), "osmosis", data))

	validators, err := database.GetValidators(context.Background(), "osmosis")
	require.NoError(t, err)
	require.Len(t, validators.Data, 2)

	data = map[string]interface{}{
		"validator1": time.Unix(0, 0),
		"validator2": time.Unix(0, 0),
		"validator3": time.Unix(0, 0),
	}
	require.NoError(t, database.StoreOrUpdateValidators(context.Background(), "osmosis", data))

	validators, err = database.GetValidators(context.Background(), "osmosis")
	require.NoError(t, err)
	require.Len(t, validators.Data, 3)
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
		Signature:        []byte("hello"),
	}
}
