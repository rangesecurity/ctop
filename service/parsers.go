package service

import (
	"fmt"
	"strconv"
	"time"

	"github.com/rangesecurity/ctop/common"
)

func parseRedisValueToNewRound(blockHeight int64, values map[string]interface{}) (*common.ParsedNewRound, error) {
	var (
		err           error
		ok            bool
		round         int64
		step          string
		proposer      string
		proposerIndex int64
	)
	if values["round"] == nil {
		return nil, fmt.Errorf("round is nil")
	} else if round_, ok := values["round"].(string); !ok {
		return nil, fmt.Errorf("failed to parse round")
	} else {
		round, err = strconv.ParseInt(round_, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("round ParseInt failed %s", err)
		}
	}

	if values["step"] == nil {
		return nil, fmt.Errorf("step is nil")
	} else if step, ok = values["step"].(string); !ok {
		return nil, fmt.Errorf("failed to parse step")
	}

	if values["proposer"] == nil {
		return nil, fmt.Errorf("proposer is nil")
	} else if proposer, ok = values["proposer"].(string); !ok {
		return nil, fmt.Errorf("failed to parse proposer")
	}

	if values["proposer_index"] == nil {
		return nil, fmt.Errorf("proposer_index is nil")
	} else if proposerIndex_, ok := values["proposer_index"].(string); !ok {
		return nil, fmt.Errorf("failed to parse proposer_index")
	} else {
		proposerIndex, err = strconv.ParseInt(proposerIndex_, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("proposer_index ParseInt failed")
		}
	}
	return &common.ParsedNewRound{
		Height:          blockHeight,
		Round:           round,
		Step:            step,
		ProposerAddress: proposer,
		ProposerIndex:   proposerIndex,
	}, nil

}

func parseRedisValueToRoundState(blockHeight int64, values map[string]interface{}) (*common.ParsedNewRoundStep, error) {
	var (
		err   error
		ok    bool
		round int64
		step  string
	)

	if values["round"] == nil {
		return nil, fmt.Errorf("round is nil")
	} else if round_, ok := values["round"].(string); !ok {
		return nil, fmt.Errorf("failed to parse round")
	} else {
		round, err = strconv.ParseInt(round_, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("round ParseInt failed %s", err)
		}
	}

	if values["step"] == nil {
		return nil, fmt.Errorf("step is nil")
	} else if step, ok = values["step"].(string); !ok {
		return nil, fmt.Errorf("failed to parse step")
	}

	return &common.ParsedNewRoundStep{
		Height: blockHeight,
		Round:  round,
		Step:   step,
	}, nil
}

func parseRedisValueToVote(blockHeight int64, values map[string]interface{}) (*common.ParsedVote, error) {
	var (
		err              error
		ok               bool
		type_            string
		round            int64
		blockID          string
		timestamp        time.Time
		validatorAddress string
		validatorIndex   int64
		signature        []byte
	)

	if values["type"] == nil {
		return nil, fmt.Errorf("type is nil")
	} else if type_, ok = values["type"].(string); !ok {
		return nil, fmt.Errorf("failed to parse type")
	}
	if values["round"] == nil {
		return nil, fmt.Errorf("round is nil")
	} else if round_, ok := values["round"].(string); !ok {
		return nil, fmt.Errorf("failed to parse round")
	} else {
		round, err = strconv.ParseInt(round_, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("round ParseInt failed %s", err)
		}

	}
	if values["block_hash"] == nil {
		return nil, fmt.Errorf("block_hash is nil")
	} else if blockID, ok = values["block_hash"].(string); !ok {
		return nil, fmt.Errorf("failed to parse block_hash")
	}
	if values["timestamp"] == nil {
		return nil, fmt.Errorf("timestamp is nil")
	} else if timestamp_, ok := values["timestamp"].(string); !ok {
		return nil, fmt.Errorf("failed to parse timestamp")
	} else {
		timestamp, err = time.Parse("2006-01-02 15:04:05.999999999 -0700 MST", timestamp_)
		if err != nil {
			return nil, fmt.Errorf("failed to parse time %v", err)
		}
	}

	if values["validator"] == nil {
		return nil, fmt.Errorf("validator is nil")
	} else if validatorAddress, ok = values["validator"].(string); !ok {
		return nil, fmt.Errorf("failed to parse validator")
	}

	if values["index"] == nil {
		return nil, fmt.Errorf("validate index is nil")
	} else if index_, ok := values["index"].(string); !ok {
		return nil, fmt.Errorf("failed to parse index")
	} else {
		validatorIndex, err = strconv.ParseInt(index_, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("failed to parse index %v", err)
		}
	}

	if values["signature"] == nil {
		return nil, fmt.Errorf("signature is nil")
	} else if signature_, ok := values["signature"].(string); !ok {
		return nil, fmt.Errorf("failed to parse signature")
	} else {
		signature = []byte(signature_)
	}

	return &common.ParsedVote{
		Type:             type_,
		Round:            round,
		Height:           blockHeight,
		BlockID:          blockID,
		Timestamp:        timestamp,
		ValidatorAddress: validatorAddress,
		ValidatorIndex:   validatorIndex,
		Signature:        signature,
	}, nil
}
