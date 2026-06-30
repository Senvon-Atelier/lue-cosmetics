-- name: CreateUser :one
INSERT INTO users (email, name, email_verified)
VALUES ($1, $2, $3)
RETURNING id, email, name, image, email_verified, created_at, updated_at;

-- name: GetUserByEmail :one
SELECT id, email, name, image, email_verified, created_at, updated_at
FROM users WHERE email = $1;

-- name: GetUserByID :one
SELECT id, email, name, image, email_verified, created_at, updated_at
FROM users WHERE id = $1;

-- name: UpdateUserEmailVerified :exec
UPDATE users SET email_verified = $2, updated_at = now() WHERE id = $1;

-- name: UpdateUserName :exec
UPDATE users SET name = $2, updated_at = now() WHERE id = $1;

-- name: UpdateUser :one
UPDATE users
SET name = COALESCE(sqlc.narg('name'), name),
    updated_at = now()
WHERE id = sqlc.arg('id')
RETURNING id, email, name, image, email_verified, created_at, updated_at;

-- name: UpsertPasswordCredential :exec
INSERT INTO password_credentials (user_id, password_hash, updated_at)
VALUES ($1, $2, now())
ON CONFLICT (user_id) DO UPDATE
SET password_hash = EXCLUDED.password_hash, updated_at = now();

-- name: GetPasswordCredentialByUserID :one
SELECT user_id, password_hash, updated_at
FROM password_credentials WHERE user_id = $1;

-- name: UpsertOAuthAccount :exec
INSERT INTO oauth_accounts (user_id, provider, provider_account_id)
VALUES ($1, $2, $3)
ON CONFLICT (provider, provider_account_id) DO NOTHING;

-- name: GetOAuthAccount :one
SELECT id, user_id, provider, provider_account_id, created_at
FROM oauth_accounts WHERE provider = $1 AND provider_account_id = $2;

-- name: CreateSession :one
INSERT INTO sessions (user_id, token_hash, expires_at, ip, user_agent)
VALUES ($1, $2, $3, $4, $5)
RETURNING id, user_id, token_hash, expires_at, ip, user_agent, created_at, last_used_at;

-- name: GetSessionByTokenHash :one
SELECT s.id, s.user_id, s.token_hash, s.expires_at, s.ip, s.user_agent, s.created_at, s.last_used_at
FROM sessions s
WHERE s.token_hash = $1 AND s.expires_at > now();

-- name: RefreshSessionLastUsed :exec
UPDATE sessions SET last_used_at = now() WHERE id = $1;

-- name: RollSessionExpiry :exec
UPDATE sessions SET expires_at = $2, last_used_at = now() WHERE id = $1;

-- name: DeleteSession :exec
DELETE FROM sessions WHERE token_hash = $1;

-- name: DeleteSessionsForUser :exec
DELETE FROM sessions WHERE user_id = $1;

-- name: DeleteOtherSessionsForUser :exec
DELETE FROM sessions WHERE user_id = $1 AND id <> $2;

-- name: CreateVerificationToken :one
INSERT INTO verification_tokens (user_id, kind, token_hash, expires_at)
VALUES ($1, $2, $3, $4)
RETURNING id, user_id, kind, token_hash, expires_at, used_at, created_at;

-- name: GetUnusedVerificationToken :one
SELECT id, user_id, kind, token_hash, expires_at, used_at, created_at
FROM verification_tokens
WHERE token_hash = $1 AND kind = $2 AND used_at IS NULL AND expires_at > now();

-- name: MarkVerificationTokenUsed :exec
UPDATE verification_tokens SET used_at = now() WHERE id = $1;

-- name: AddUserRole :exec
INSERT INTO user_roles (user_id, role) VALUES ($1, $2) ON CONFLICT DO NOTHING;

-- name: ListRolesForUser :many
SELECT role FROM user_roles WHERE user_id = $1 ORDER BY role;
