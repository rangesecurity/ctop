CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

--bun:split

CREATE TABLE vote_events (
    id UUID NOT NULL PRIMARY KEY DEFAULT uuid_generate_v4(),
    network TEXT NOT NULL,
    vote_type TEXT NOT NULL,
    height INTEGER NOT NULL,
    round INTEGER NOT NULL,
    block_id TEXT NOT NULL,
    block_timestamp TIMESTAMPTZ NOT NULL,
    validator_address TEXT NOT NULL,
    validator_index TEXT NOT NULL,
    validator_signature BYTEA NOT NULL
);