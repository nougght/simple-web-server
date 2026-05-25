CREATE TABLE IF NOT EXISTS notes(
    note_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    header VARCHAR(256) NOT NULL,
    body TEXT
)