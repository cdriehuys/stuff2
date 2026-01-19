-- Create a trigger to manage the `updated_at` column for a table.
CREATE FUNCTION _manage_updated_at(_tbl regclass) RETURNS VOID AS $$
BEGIN
    EXECUTE format('CREATE TRIGGER set_updated_at BEFORE UPDATE ON %s
                    FOR EACH ROW EXECUTE PROCEDURE _set_updated_at()', _tbl);
END;
$$ LANGUAGE plpgsql;

-- Update the `updated_at` column if data changed.
CREATE FUNCTION _set_updated_at() RETURNS trigger AS $$
BEGIN
    IF (
        NEW IS DISTINCT FROM OLD AND
        NEW.updated_at IS NOT DISTINCT FROM OLD.updated_at
    ) THEN
        NEW.updated_at := current_timestamp;
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TABLE users(
    id uuid PRIMARY KEY,
    email TEXT NOT NULL,
    email_verified_at TIMESTAMPTZ,
    password_hash TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

SELECT _manage_updated_at('users');

CREATE UNIQUE INDEX users_email_verified_key ON users(email)
    WHERE email_verified_at IS NOT NULL;

CREATE TABLE email_verification_keys(
    id INT PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    user_id uuid NOT NULL REFERENCES users(id)
        ON DELETE CASCADE,
    email TEXT NOT NULL,
    token TEXT UNIQUE NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

---- create above / drop below ----

DROP TABLE email_verification_keys;

DROP INDEX users_email_verified_key;
DROP TABLE users;

DROP FUNCTION _manage_updated_at;
DROP FUNCTION _set_updated_at;
