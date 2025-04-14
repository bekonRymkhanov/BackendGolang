
CREATE TABLE IF NOT EXISTS user_favorite_books (
    id SERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL,
    book_name VARCHAR(255) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE (user_id, book_name),
    CONSTRAINT fk_user_id FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);