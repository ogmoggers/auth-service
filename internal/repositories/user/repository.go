package user

import (
	"context"
	"database/sql"
	"errors"

	"auth-service/pkg/domain"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{
		db: db,
	}
}

func (r *Repository) Save(ctx context.Context, user *domain.User) error {
	query := `
INSERT INTO users (id, email, password, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5)
`

	_, err := r.db.ExecContext(
		ctx,
		query,
		user.ID,
		user.Email,
		user.Password,
		user.CreatedAt,
		user.UpdatedAt,
	)

	return err
}

func (r *Repository) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	query := `
SELECT id, email, password, created_at, updated_at
FROM users
WHERE email = $1
`

	var user domain.User
	err := r.db.QueryRowContext(ctx, query, email).Scan(
		&user.ID,
		&user.Email,
		&user.Password,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, errors.New("user not found")
	}

	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *Repository) SaveAuthCode(ctx context.Context, authCode *domain.AuthCode) error {
	query := `
INSERT INTO auth_codes (user_id, code, expires_at)
VALUES ($1, $2, $3)
ON CONFLICT (user_id, code) DO UPDATE
SET expires_at = $3
`

	_, err := r.db.ExecContext(
		ctx,
		query,
		authCode.UserID,
		authCode.Code,
		authCode.ExpiresAt,
	)

	return err
}

func (r *Repository) GetAuthCode(ctx context.Context, userID string) (*domain.AuthCode, error) {
	query := `
SELECT user_id, code, expires_at
FROM auth_codes
WHERE user_id = $1
ORDER BY expires_at DESC
LIMIT 1
`

	var authCode domain.AuthCode
	err := r.db.QueryRowContext(ctx, query, userID).Scan(
		&authCode.UserID,
		&authCode.Code,
		&authCode.ExpiresAt,
	)

	if err == sql.ErrNoRows {
		return nil, errors.New("auth code not found")
	}

	if err != nil {
		return nil, err
	}

	return &authCode, nil
}
