package db

import (
	"time"

	"github.com/google/uuid"
)

type VoteEvent struct {
	// tableName is an optional field that specifies custom table name and alias.
	// By default go-pg generates table name and alias from struct name.
	//lint:ignore U1000 Ignore
	tableName struct{} `pg:"vote_events"`

	ID               uuid.UUID `pg:"type:uuid,default:gen_random_uuid()"`
	Network          string
	Type             string
	Height           int
	Round            int    `pg:"default:0"`
	BlockID          string `pb:"block_id"`
	Timestamp        time.Time
	ValidatorAddress string `pg:"validator_address"`
	ValidatorIndex   int    `pg:"validator_index"`
	Signature        []byte
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
