package store

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"

	"github.com/zagvozdeen/again/internal/models"
)

var ErrUserNotFound = errors.New("user not found")

func (s *Store) GetUserByAPIToken(ctx context.Context, token string) (*models.User, error) {
	hash := sha256.Sum256([]byte(token))
	u := &models.User{}
	err := s.querier(ctx).QueryRow(
		ctx,
		"WITH used AS (UPDATE api_tokens SET last_used_at = now() WHERE token_hash = $1 RETURNING user_id) "+
			"SELECT u.id, u.email, u.timezone, u.ui_language, u.created_at FROM users u JOIN used ON used.user_id = u.id",
		hex.EncodeToString(hash[:]),
	).Scan(&u.ID, &u.Email, &u.Timezone, &u.UILanguage, &u.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to get user by api token: %w", err)
	}
	return u, nil
}
