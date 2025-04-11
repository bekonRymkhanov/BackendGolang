CREATE TABLE genres (
    id SERIAL PRIMARY KEY,
    title TEXT NOT NULL,
    subgenre_count INTEGER NOT NULL,
    url TEXT NOT NULL,
    version integer NOT NULL DEFAULT 1

);