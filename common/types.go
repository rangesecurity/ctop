package common

import "time"

// wrapper around types.Vote which provides easier to store types
type ParsedVote struct {
	Type             string
	Round            int64
	Height           int64
	BlockID          string
	Timestamp        time.Time
	ValidatorAddress string
	ValidatorIndex   int64
	Signature        []byte
}

// wrapper around types.EventDataNewRound which provides easier to store types
type ParsedNewRound struct {
	Height          int64
	Round           int64
	Step            string
	ProposerAddress string
	ProposerIndex   int64
}

// wrapper around types.EventDataRoundState which provides easier to store types
type ParsedNewRoundStep struct {
	Height int64
	Round  int64
	Step   string
}
