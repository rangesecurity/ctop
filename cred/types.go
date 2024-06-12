package cred

import "github.com/redis/go-redis/v9"

var VotesLuaScript = redis.NewScript(`
local base_key = ARGV[1]
local block_height = ARGV[2]
local validator = ARGV[3]
local round = ARGV[4]
local signature = ARGV[5]
local index = ARGV[6]
local block_hash = ARGV[7]
local vote_type = ARGV[8]
local timestamp = ARGV[9]

-- Generate the sequence key based on the base key and block height
local sequence_key = base_key .. ":sequence_vote:" .. block_height

-- Increment the sequence counter
local sequence = redis.call("INCR", sequence_key)

-- Construct the ID
local id = block_height .. "-" .. sequence

-- Add entry to the stream
redis.call("XADD", base_key .. ":votes", id, "validator", validator, "round", round, "signature", signature, "index", index, "block_hash", block_hash, "type", vote_type, "timestamp", timestamp)

return id
`)

var NewRoundStepScript = redis.NewScript(`
local base_key = ARGV[1]
local block_height = ARGV[2]
local round = ARGV[3]
local step = ARGV[4]

-- Generate the sequence key based on the base key and block height
local sequence_key = base_key .. ":sequence_round_step:" .. block_height

-- Increment the sequence counter
local sequence = redis.call("INCR", sequence_key)

-- Construct the ID
local id = block_height .. "-" .. sequence

-- Add entry to the stream
redis.call("XADD", base_key .. ":new_round_step", id, "round", round, "step", step)

return id
`)

var NewRoundScript = redis.NewScript(`
local base_key = ARGV[1]
local block_height = ARGV[2]
local round = ARGV[3]
local step = ARGV[4]
local proposer = ARGV[5]
local proposer_index = ARGV[6]

-- Generate the sequence key based on the base key and block height
local sequence_key = base_key .. ":sequence_round:" .. block_height

-- Increment the sequence counter
local sequence = redis.call("INCR", sequence_key)

-- Construct the ID
local id = block_height .. "-" .. sequence

-- Add entry to the stream
redis.call("XADD", base_key .. ":new_round", id, 
           "round", round, 
           "step", step, 
           "proposer", proposer, 
           "proposer_index", proposer_index)

return id
`)
