ALTER TABLE mood_entries
ADD COLUMN IF NOT EXISTS auth0_user_id TEXT REFERENCES users(auth0_user_id);

INSERT INTO users (auth0_user_id)
VALUES ('auth0|68164b4c821b56fdc024b2dd')
ON CONFLICT (auth0_user_id) DO NOTHING;

UPDATE mood_entries
SET auth0_user_id = 'auth0|68164b4c821b56fdc024b2dd'
WHERE auth0_user_id IS NULL;

CREATE INDEX IF NOT EXISTS mood_entries_auth0_user_id_idx
ON mood_entries (auth0_user_id);

ALTER TABLE mood_entries
ALTER COLUMN auth0_user_id SET NOT NULL;
