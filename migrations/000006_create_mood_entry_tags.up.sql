CREATE TABLE mood_entry_tags (
  mood_entry_id UUID NOT NULL REFERENCES mood_entries(id) ON DELETE CASCADE,
  mood_tag_id UUID NOT NULL REFERENCES mood_tags(id) ON DELETE RESTRICT,
  PRIMARY KEY (mood_entry_id, mood_tag_id)
);

CREATE INDEX mood_entry_tags_mood_tag_id_idx ON mood_entry_tags (mood_tag_id);
