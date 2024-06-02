package db

import (
	"github.com/cometbft/cometbft/types"
	"github.com/go-pg/pg/v10"
	"github.com/go-pg/pg/v10/orm"
)

type Database struct {
	DB *pg.DB
}

func New(url string) (*Database, error) {
	opt, err := pg.ParseURL(url)
	if err != nil {
		return nil, err
	}
	opt.TLSConfig = nil
	db := &Database{DB: pg.Connect(opt)}
	return db, db.CreateSchema()
}

func (d *Database) StoreVote(
	network string,
	vote types.Vote,
) error {
	_, err := d.DB.Model(&VoteEvent{
		Network:          network,
		Type:             vote.Type.String(),
		Height:           vote.Height,
		Round:            vote.Round,
		BlockID:          vote.BlockID.String(),
		Timestamp:        vote.Timestamp,
		ValidatorAddress: vote.ValidatorAddress.String(),
		ValidatorIndex:   vote.ValidatorIndex,
		Signature:        vote.Signature,
	}).Insert()
	return err
}

func (d *Database) StoreNewRound(
	network string,
	roundInfo types.EventDataNewRound,
) error {
	_, err := d.DB.Model(&NewRoundEvent{
		Network:          network,
		Height:           roundInfo.Height,
		Round:            roundInfo.Round,
		Step:             roundInfo.Step,
		ValidatorAddress: roundInfo.Proposer.Address.String(),
		ValidatorIndex:   roundInfo.Proposer.Index,
	}).Insert()
	return err
}

func (d *Database) StoreNewRoundStep(
	network string,
	roundInfo types.EventDataRoundState,
) error {
	_, err := d.DB.Model(&NewRoundStepEvent{
		Network: network,
		Height:  roundInfo.Height,
		Round:   roundInfo.Round,
		Step:    roundInfo.Step,
	}).Insert()
	return err
}
func (d *Database) GetVotes(network string) (votes []VoteEvent, err error) {
	err = d.DB.Model(&votes).Where("network = ?", network).Select()
	return
}

func (d *Database) GetNewRounds(network string) (rounds []NewRoundEvent, err error) {
	err = d.DB.Model(&rounds).Where("network = ?", network).Select()
	return
}

func (d *Database) GetNewRoundSteps(network string) (steps []NewRoundStepEvent, err error) {
	err = d.DB.Model(&steps).Where("network = ?", network).Select()
	return
}

func (d *Database) CreateSchema() error {
	models := []interface{}{
		(*VoteEvent)(nil),
		(*NewRoundEvent)(nil),
		(*NewRoundStepEvent)(nil),
	}

	for _, model := range models {
		err := d.DB.Model(model).CreateTable(&orm.CreateTableOptions{
			IfNotExists: true,
		})
		if err != nil {
			return err
		}
	}
	return nil
}
