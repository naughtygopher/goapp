CREATE TABLE IF NOT EXISTS Users (
    id BIGSERIAL PRIMARY KEY,
    firstName TEXT,
    lastName TEXT,
    mobile TEXT,
    email TEXT UNIQUE,
    createdAt timestamptz DEFAULT now(),
    updatedAt timestamptz DEFAULT now()
);
