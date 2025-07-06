-- Create posts table
CREATE TABLE IF NOT EXISTS posts (
    id TEXT PRIMARY KEY,
    title TEXT NOT NULL,
    content TEXT NOT NULL,
    auth0_user_id TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    slug TEXT UNIQUE NOT NULL
);

-- Create index on slug for faster lookups
CREATE INDEX IF NOT EXISTS idx_posts_slug ON posts(slug);

-- Create index on auth0_user_id for faster lookups
CREATE INDEX IF NOT EXISTS idx_posts_auth0_user_id ON posts(auth0_user_id); 