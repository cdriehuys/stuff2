-- The structure of this table is dictated by the `scs/pgxstore` module as
-- described here:
-- https://github.com/alexedwards/scs/tree/master/pgxstore#setup

CREATE TABLE sessions (
	token TEXT PRIMARY KEY,
	data BYTEA NOT NULL,
	expiry TIMESTAMPTZ NOT NULL
);

CREATE INDEX sessions_expiry_idx ON sessions (expiry);

---- create above / drop below ----

DROP TABLE sessions;
