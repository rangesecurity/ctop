CREATE TABLE new_round_step_events (
    id UUID NOT NULL PRIMARY KEY DEFAULT uuid_generate_v4(),
    network TEXT NOT NULL,
    height INTEGER NOT NULL,
    round INTEGER NOT NULL,
    step TEXT NOT NULL
);