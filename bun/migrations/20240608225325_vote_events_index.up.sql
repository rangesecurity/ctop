CREATE INDEX idx_vote_events_validator_network_height ON vote_events (validator_address, network, height);

CREATE INDEX idx_vote_events_network ON vote_events (network);