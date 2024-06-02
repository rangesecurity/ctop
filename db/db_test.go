package db_test

import (
	"testing"
	"time"

	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/go-pg/pg/v10/orm"

	"github.com/cometbft/cometbft/crypto"
	"github.com/cometbft/cometbft/crypto/tmhash"
	"github.com/cometbft/cometbft/types"
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
		}
		for _, model := range models {
			database.DB.Model(model).DropTable(&orm.DropTableOptions{
				IfExists: true,
			})
		}
	}
	require.NoError(t, database.StoreVote("osmosis", *exampleVote(12345, byte(cmtproto.PrevoteType))))
	votes, err := database.GetVotes("osmosis")
	require.NoError(t, err)
	require.Len(t, votes, 1)

	mockKey1 := types.NewMockPV()
	validator1 := types.NewValidator(mockKey1.PrivKey.PubKey(), 10)
	require.NoError(t, database.StoreNewRound("osmosis", types.EventDataNewRound{
		Height: 112345,
		Round:  0,
		Step:   "step",
		Proposer: types.ValidatorInfo{
			Address: validator1.Address,
			Index:   1,
		},
	}))

	newRounds, err := database.GetNewRounds("osmosis")
	require.NoError(t, err)
	require.Len(t, newRounds, 1)

	require.NoError(t, database.StoreNewRoundStep(
		"osmosis",
		types.EventDataRoundState{
			Height: 11234,
			Round:  0,
			Step:   "RoundStepPropose",
		},
	))
	roundSteps, err := database.GetNewRoundSteps("osmosis")
	require.NoError(t, err)
	require.Len(t, roundSteps, 1)
	cleanUp()
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
