package common

import "time"

// wrapper around types.Vote which provides easier to store types
type ParsedVote struct {
	Type             string
	Round            int64
	BlockNumber      int64
	BlockID          string
	Timestamp        time.Time
	ValidatorAddress string
	ValidatorIndex   int64
	Signature        []byte
}
