package db

import (
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type VoteEvent struct {
	bun.BaseModel      `bun:"table:vote_events"`
	ID                 uuid.UUID
	Network            string
	VoteType           string
	Height             int
	Round              int
	BlockID            string
	BlockTimestamp     time.Time
	ValidatorAddress   string `pg:"validator_address"`
	ValidatorIndex     int    `pg:"validator_index"`
	ValidatorSignature []byte
}

type NewRoundEvent struct {
	//lint:ignore U1000 Ignore
	tableName struct{} `pg:"new_round_events"`

	ID      uuid.UUID `pg:"type:uuid,default:gen_random_uuid()"`
	Network string

	Height           int
	Round            int
	Step             string
	ValidatorAddress string `pg:"validator_address"`
	ValidatorIndex   int    `pg:"validator_index"`
}

type NewRoundStepEvent struct {
	//lint:ignore U1000 Ignore
	tableName struct{} `pg:"new_round_step_events"`

	ID      uuid.UUID `pg:"type:uuid,default:gen_random_uuid()"`
	Network string

	Height int
	Round  int
	Step   string
}
