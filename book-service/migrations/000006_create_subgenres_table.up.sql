CREATE TABLE subgenres (
    id SERIAL PRIMARY KEY,
    title TEXT NOT NULL,
    main_genre TEXT NOT NULL,
    book_count NUMERIC NOT NULL,
    url TEXT NOT NULL,
    version integer NOT NULL DEFAULT 1
);