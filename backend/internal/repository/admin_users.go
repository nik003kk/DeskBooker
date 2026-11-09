package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/nik003kk/DeskBooker/backend/internal/models"
)

// Explicit safe columns keep password hashes out of this read path entirely.
func (s *Store) ListAdminUsers(ctx context.Context, limit, offset int) (models.AdminUserPage, error) {
	if limit < 1 || limit > 100 || offset < 0 || offset > 1000000 {
		return models.AdminUserPage{}, errors.New("invalid user pagination")
	}
	const query = `SELECT jsonb_build_object(
 'users',COALESCE((SELECT jsonb_agg(to_jsonb(p) ORDER BY lower(p.email),p.id) FROM (
 SELECT id,email,first_name,last_name,role,status,department,created_at
 FROM public.users ORDER BY lower(email),id LIMIT $1 OFFSET $2) p),'[]'::jsonb),
 'total',(SELECT count(*) FROM public.users))`
	var raw []byte
	if err := s.db.QueryRow(ctx, query, limit, offset).Scan(&raw); err != nil {
		return models.AdminUserPage{}, fmt.Errorf("list admin users: %w", err)
	}
	var page models.AdminUserPage
	if err := json.Unmarshal(raw, &page); err != nil {
		return models.AdminUserPage{}, fmt.Errorf("decode admin users: %w", err)
	}
	page.Limit = limit
	page.Offset = offset
	return page, nil
}
