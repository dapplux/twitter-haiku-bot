CREATE TABLE posts (
    id TEXT PRIMARY KEY,
    author JSONB NOT NULL,
    text TEXT NOT NULL,
    likes INT DEFAULT 0,
    shares INT DEFAULT 0,
    replies INT DEFAULT 0,
    platform TEXT NOT NULL CHECK (platform IN ('twitter')),
    created_at TIMESTAMP DEFAULT now()
);

CREATE TYPE haiku_state AS ENUM (
    'created',
    'summary_getting',
    'summary_got',
    'haiku_text_getting',
    'haiku_text_got',
    'comenting',
    'done',
    'failed'
);

CREATE TYPE platform AS ENUM (
    'twitter'
);

CREATE TABLE haikus (
    id TEXT PRIMARY KEY,
    state haiku_state NOT NULL DEFAULT 'created',
    summary TEXT,
    text TEXT,
    post_id TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT now(),
    updated_at TIMESTAMP DEFAULT now(),
    FOREIGN KEY (post_id) REFERENCES posts(id)
);

CREATE INDEX idx_posts_created_at ON posts(created_at);
CREATE INDEX idx_haikus_post_id ON haikus(post_id);
