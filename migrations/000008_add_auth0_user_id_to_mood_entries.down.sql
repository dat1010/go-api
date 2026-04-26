DROP INDEX IF EXISTS mood_entries_auth0_user_id_idx;

ALTER TABLE mood_entries
DROP COLUMN IF EXISTS auth0_user_id;
