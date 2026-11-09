package repository

import (
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/nik003kk/DeskBooker/backend/internal/models"
	"strings"
)

var (
	ErrAdminRequired  = errors.New("active administrator required")
	ErrSelfRoleChange = errors.New("cannot change own role")
)

type UserAdminStore struct{ db TxStarter }

func NewUserAdminStore(db TxStarter) *UserAdminStore { return &UserAdminStore{db: db} }

func (s *UserAdminStore) ChangeUserRole(ctx context.Context, actorID, targetID string, role models.UserRole) (models.UserDTO, error) {
	if strings.EqualFold(actorID, targetID) {
		return models.UserDTO{}, ErrSelfRoleChange
	}
	if canonical, ok := models.ParseUserRole(string(role)); !ok || canonical != role {
		return models.UserDTO{}, errors.New("invalid persisted role")
	}
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return models.UserDTO{}, fmt.Errorf("begin role change: %w", err)
	}
	defer tx.Rollback(ctx)
	// Deterministic lock order prevents cross-administrator deadlocks. The actor
	// must still be an active admin after waiting for any concurrent role change.
	rows, err := tx.Query(ctx, "SELECT id,role,status FROM public.users WHERE id=ANY($1::uuid[]) ORDER BY id FOR UPDATE", []string{actorID, targetID})
	if err != nil {
		return models.UserDTO{}, fmt.Errorf("lock role change: %w", err)
	}
	var actor models.User
	found := false
	for rows.Next() {
		var u models.User
		if err = rows.Scan(&u.ID, &u.Role, &u.Status); err != nil {
			rows.Close()
			return models.UserDTO{}, err
		}
		if strings.EqualFold(u.ID, actorID) {
			actor = u
		}
		if strings.EqualFold(u.ID, targetID) {
			found = true
		}
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return models.UserDTO{}, err
	}
	if actor.Role != models.RoleAdmin || actor.Status != models.UserActive {
		return models.UserDTO{}, ErrAdminRequired
	}
	if !found {
		return models.UserDTO{}, pgx.ErrNoRows
	}
	var saved models.UserDTO
	err = tx.QueryRow(ctx, "UPDATE public.users SET role=$2 WHERE id=$1 RETURNING id,email,first_name,last_name,role,status,department,created_at", targetID, role).Scan(&saved.ID, &saved.Email, &saved.FirstName, &saved.LastName, &saved.Role, &saved.Status, &saved.Department, &saved.CreatedAt)
	if err != nil {
		return models.UserDTO{}, fmt.Errorf("update user role: %w", err)
	}
	if err = tx.Commit(ctx); err != nil {
		return models.UserDTO{}, fmt.Errorf("commit user role: %w", err)
	}
	return saved, nil
}
