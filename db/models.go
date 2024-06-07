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
	ValidatorAddress   string
	ValidatorIndex     int
	ValidatorSignature []byte
}

type NewRoundEvent struct {
	bun.BaseModel `bun:"table:new_round_events"`

	ID      uuid.UUID
	Network string

	Height           int
	Round            int
	Step             string
	ValidatorAddress string
	ValidatorIndex   int
}

type NewRoundStepEvent struct {
	bun.BaseModel `bun:"table:new_round_step_events"`

	ID      uuid.UUID
	Network string

	Height int
	Round  int
	Step   string
}

type Validators struct {
	bun.BaseModel `bun:"table:validators"`
	ID            uuid.UUID
	Network       string
	// map validator_address => last_vote_time
	Data map[string]interface{} `bun:"type:jsonb"`
}
