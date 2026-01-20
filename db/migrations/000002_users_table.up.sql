-- user status enum
CREATE TYPE user_status AS ENUM ('active', 'deleted');

-- users table
CREATE TABLE IF NOT EXISTS users
(
    uuid uuid NOT NULL
        CONSTRAINT user_uuid_pkey PRIMARY KEY,
    email text,
    name text,
    status user_status NOT NULL DEFAULT 'active',
    created_at timestamp without time zone NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp without time zone NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS user_status_idx
    ON users USING btree(status);

CREATE UNIQUE INDEX IF NOT EXISTS user_email_idx
    ON users USING btree(email);

-- update trigger
CREATE TRIGGER update_user_updated_at
    BEFORE UPDATE
    ON users
    FOR EACH ROW
EXECUTE PROCEDURE update_updated_at_column();
