CREATE TABLE IF NOT EXISTS user_invitation (
    token bytea,
    user_id bigint NOT NULL
    REFERENCES users,
    PRIMARY KEY (token, user_id)
);

ALTER TABLE
  users
ADD
  COLUMN is_active BOOLEAN NOT NULL DEFAULT FALSE;