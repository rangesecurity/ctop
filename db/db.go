package db

import (
	"context"
	"database/sql"

	"github.com/cometbft/cometbft/types"
	"github.com/rangesecurity/ctop/cmd/bun/migrations"
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
	vote types.Vote,
) error {
	_, err := d.DB.NewInsert().Model(&VoteEvent{
		Network:            network,
		VoteType:           vote.Type.String(),
		Height:             int(vote.Height),
		Round:              int(vote.Round),
		BlockID:            vote.BlockID.String(),
		BlockTimestamp:     vote.Timestamp,
		ValidatorAddress:   vote.ValidatorAddress.String(),
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
