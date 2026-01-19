-- name: DeleteEmailVerificationKeyByID :exec
DELETE FROM email_verification_keys
WHERE id = @id;

-- name: DeleteUnverifiedEmails :exec
DELETE FROM users
WHERE email = @email AND email_verified_at IS NULL;

-- name: GetEmailVerificationKeyByToken :one
SELECT * FROM email_verification_keys
WHERE token = @token;

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
    WHERE email = @email and email_verified_at IS NOT NULL
);

-- name: VerifyEmailForUser :exec
UPDATE users
SET email_verified_at = now()
WHERE id = @user_id;
