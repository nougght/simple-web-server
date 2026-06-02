CREATE TABLE notes(
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    header VARCHAR(256) NOT NULL,
    body TEXT
)