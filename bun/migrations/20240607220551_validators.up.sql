CREATE TABLE validators (
    id UUID NOT NULL PRIMARY KEY DEFAULT uuid_generate_v4(),
    network TEXT NOT NULL UNIQUE,
    data JSONB NOT NULL
);