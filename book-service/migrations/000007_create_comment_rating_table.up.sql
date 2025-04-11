-- Create comments table
CREATE TABLE IF NOT EXISTS comments (
    id bigserial PRIMARY KEY,
    book_id bigint NOT NULL REFERENCES books(id) ON DELETE CASCADE,
    user_id bigint NOT NULL,
    content text NOT NULL,
    created_at timestamp(0) with time zone NOT NULL DEFAULT NOW(),
    version integer NOT NULL DEFAULT 1
);

CREATE INDEX IF NOT EXISTS comments_book_id_idx ON comments(book_id);
CREATE INDEX IF NOT EXISTS comments_user_id_idx ON comments(user_id);

-- Create ratings table
CREATE TABLE IF NOT EXISTS ratings (
    id bigserial PRIMARY KEY,
    book_id bigint NOT NULL REFERENCES books(id) ON DELETE CASCADE,
    user_id bigint NOT NULL,
    score integer NOT NULL CHECK (score >= 1 AND score <= 5),
    created_at timestamp(0) with time zone NOT NULL DEFAULT NOW(),
    version integer NOT NULL DEFAULT 1,
    -- Ensure a user can only rate a book once
    UNIQUE (book_id, user_id)
);

CREATE INDEX IF NOT EXISTS ratings_book_id_idx ON ratings(book_id);
CREATE INDEX IF NOT EXISTS ratings_user_id_idx ON ratings(user_id);