package db

import (
	"context"
	"database/sql"

	"github.com/cometbft/cometbft/types"
	"github.com/rangesecurity/ctop/bun/migrations"
	"github.com/rangesecurity/ctop/common"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
	"github.com/uptrace/bun/extra/bundebug"
	"github.com/uptrace/bun/migrate"
)

type Database struct {
	DB *bun.DB
}

func New(url string) (*Database, error) {
	db := OpenDB(url)
	return &Database{DB: db}, nil
}

func (d *Database) StoreVote(
	ctx context.Context,
	network string,
	vote common.ParsedVote,
) error {
	_, err := d.DB.NewInsert().Model(&VoteEvent{
		Network:            network,
		VoteType:           vote.Type,
		Height:             int(vote.BlockNumber),
		Round:              int(vote.Round),
		BlockID:            vote.BlockID,
		BlockTimestamp:     vote.Timestamp,
		ValidatorAddress:   vote.ValidatorAddress,
		ValidatorIndex:     int(vote.ValidatorIndex),
		ValidatorSignature: vote.Signature,
	}).Exec(ctx)
	return err
}

func (d *Database) StoreNewRound(
	ctx context.Context,
	network string,
	roundInfo types.EventDataNewRound,
) error {
	_, err := d.DB.NewInsert().Model(&NewRoundEvent{
		Network:          network,
		Height:           int(roundInfo.Height),
		Round:            int(roundInfo.Round),
		Step:             roundInfo.Step,
		ValidatorAddress: roundInfo.Proposer.Address.String(),
		ValidatorIndex:   int(roundInfo.Proposer.Index),
	}).Exec(ctx)
	return err
}

func (d *Database) StoreNewRoundStep(
	ctx context.Context,
	network string,
	roundInfo types.EventDataRoundState,
) error {
	_, err := d.DB.NewInsert().Model(&NewRoundStepEvent{
		Network: network,
		Height:  int(roundInfo.Height),
		Round:   int(roundInfo.Round),
		Step:    roundInfo.Step,
	}).Exec(ctx)
	return err
}

func (d *Database) StoreOrUpdateValidators(
	ctx context.Context,
	network string,
	data map[string]interface{},
) error {
	return d.DB.RunInTx(ctx, &sql.TxOptions{}, func(ctx context.Context, tx bun.Tx) error {
		var (
			validators Validators
			err        error
		)
		if err = tx.NewSelect().Model(&validators).Where("network = ?", network).Scan(ctx); err != nil {
			validators = Validators{
				Network: network,
				Data:    data,
			}
			_, err = tx.NewInsert().Model(&validators).Exec(ctx)
		} else {
			if validators.Data == nil {
				validators.Data = make(map[string]interface{})
			}
			for k, v := range data {
				validators.Data[k] = v
			}
			_, err = tx.NewUpdate().Model(&validators).Column("data").Where("network = ?", network).Exec(ctx)
		}
		return err
	})
}

func (d *Database) GetVotes(ctx context.Context, network string) (votes []VoteEvent, err error) {
	err = d.DB.NewSelect().Model(&votes).Where("network = ?", network).Scan(ctx)
	return
}

func (d *Database) GetNewRounds(ctx context.Context, network string) (rounds []NewRoundEvent, err error) {
	err = d.DB.NewSelect().Model(&rounds).Where("network = ?", network).Scan(ctx)
	return
}

func (d *Database) GetNewRoundSteps(ctx context.Context, network string) (steps []NewRoundStepEvent, err error) {
	err = d.DB.NewSelect().Model(&steps).Where("network = ?", network).Scan(ctx)
	return
}

func (d *Database) GetValidators(ctx context.Context, network string) (validators Validators, err error) {
	err = d.DB.NewSelect().Model(&validators).Where("network = ?", network).Scan(ctx)
	return
}

func (d *Database) GetLatestVotesForNetwork(
	ctx context.Context,
	network string,
) ([]VoteEvent, error) {
	subquery := d.DB.NewSelect().
		Model((*VoteEvent)(nil)).
		Column("validator_address", "network").
		ColumnExpr("MAX(height) AS max_height").
		Where("network IN (?)", bun.In([]string{network})).
		Group("validator_address", "network")

	var voteEvents []VoteEvent
	err := d.DB.NewSelect().
		Model(&voteEvents).
		With("latest_heights", subquery).
		TableExpr("vote_events AS ve").
		Join("JOIN latest_heights AS lh ON ve.validator_address = lh.validator_address AND ve.network = lh.network AND ve.height = lh.max_height").
		Where("ve.network IN (?)", bun.In([]string{network})).
		Scan(ctx)
	return voteEvents, err
}

func (d *Database) CreateSchema(ctx context.Context) error {
	migrator := migrate.NewMigrator(d.DB, migrations.Migrations)
	_, err := migrator.Migrate(ctx)
	return err
}

func OpenDB(url string) *bun.DB {
	sqldb := sql.OpenDB(pgdriver.NewConnector(
		pgdriver.WithDSN(url),
		pgdriver.WithTLSConfig(nil),
	))
	db := bun.NewDB(sqldb, pgdialect.New())
	db.AddQueryHook(bundebug.NewQueryHook(
		bundebug.WithEnabled(false),
		bundebug.FromEnv(""),
	))
	return db
}
