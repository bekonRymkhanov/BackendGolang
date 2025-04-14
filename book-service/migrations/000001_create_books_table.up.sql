CREATE TABLE IF NOT EXISTS books (
    id SERIAL PRIMARY KEY,
    title TEXT NOT NULL,
    author TEXT,
    main_genre TEXT,
    sub_genre TEXT,
    type TEXT,
    price TEXT,
    rating REAL,
    people_rated BIGINT,
    url TEXT,
    version BIGINT
);
