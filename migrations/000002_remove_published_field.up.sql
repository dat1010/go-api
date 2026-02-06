-- Remove the published column from posts table
-- Since all posts are always published for authenticated users, this field is redundant

ALTER TABLE posts DROP COLUMN published; 