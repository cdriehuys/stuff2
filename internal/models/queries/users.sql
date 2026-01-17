-- name: InsertEmailVerificationKey :exec
INSERT INTO email_verification_keys(user_id, email, token)
VALUES (@user_id, @email, @token);

-- name: InsertNewUser :one
INSERT INTO users (id, email, password_hash)
VALUES (@id, @email, @password_hash)
RETURNING *;

-- name: VerifiedEmailExists :one
SELECT EXISTS(
    SELECT 1 FROM users
    WHERE email = @email and email_verified
);
