
CREATE EXTENSION IF NOT EXISTS citext;

CREATE TABLE IF NOT EXISTS users (
    "id" SERIAL PRIMARY KEY,
    "username" TEXT NOT NULL,
    "email" CITEXT NOT NULL UNIQUE, -- 比较时忽略大小写
    "password_hash" TEXT NOT NULL, 
    "created_at" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updated_at" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_email ON users(email);


CREATE TABLE IF NOT EXISTS urls (
    "id" BIGSERIAL PRIMARY KEY,
    "user_id" INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    "original_url" TEXT NOT NULL,
    "short_code" TEXT NOT NULL UNIQUE,
    "is_custom" BOOLEAN NOT NULL DEFAULT FALSE,
    "views" INT NOT NULL DEFAULT 0,
    "expired_at" TIMESTAMP NOT NULL,
    "created_at" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_short_code ON urls(short_code);
CREATE INDEX idx_expired_at_short_code ON urls(short_code, expired_at); -- 设置复合索引

CREATE INDEX idx_user_id ON urls(user_id);
CREATE INDEX idx_expired_at_user_id ON urls(user_id, expired_at); -- 设置复合索引