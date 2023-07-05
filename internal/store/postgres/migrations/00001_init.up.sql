BEGIN TRANSACTION;

CREATE TABLE shortener(
    slug VARCHAR(255),
    original_url VARCHAR(255) PRIMARY KEY,
    user_id VARCHAR(255),
    deleted_flag BOOLEAN DEFAULT FALSE,
    UNIQUE(slug, original_url)
);

COMMIT;