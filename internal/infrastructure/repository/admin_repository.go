package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/tharunn0/E-Commerce-Go/internal/domain"
)

type AdminRepository struct {
	DB *pgxpool.Pool
}

func NewAdminRepository(db *pgxpool.Pool) *AdminRepository {
	return &AdminRepository{
		DB: db,
	}
}

func (repo *AdminRepository) GetAdmin(ctx context.Context, email string) (*domain.User, error) {
	user := &domain.User{}
	err := repo.DB.QueryRow(ctx,
		`SELECT id, email, phone, password, first_name, last_name, role,is_verified , status FROM users
    WHERE role = 'admin' AND email = $1;
	`, email).Scan(&user.ID, &user.Email, &user.Phone, &user.Password, &user.FirstName,
		&user.LastName, &user.Role, &user.IsVerified, &user.Status)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (repo *AdminRepository) GetUserByEmail(ctx context.Context, email string) (*domain.User, error) {
	user := &domain.User{}
	err := repo.DB.QueryRow(ctx,
		`SELECT id, email, phone, password, first_name, last_name, role, is_verified, status, created_at, updated_at
		 FROM users WHERE email = $1 AND status = 'active'`,
		email).Scan(&user.ID, &user.Email, &user.Phone, &user.Password, &user.FirstName,
		&user.LastName, &user.Role, &user.IsVerified, &user.Status, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (repo *AdminRepository) GetUsersByID(ctx context.Context, id int64) (*domain.UserProfile, error) {

	user := &domain.UserProfile{}

	query := `SELECT id, email,COALESCE(phone,''), first_name, last_name, role, is_verified, status, created_at, updated_at FROM users
	WHERE id = $1`

	err := repo.DB.QueryRow(ctx, query, id).Scan(&user.ID, &user.Email, &user.Phone, &user.FirstName,
		&user.LastName, &user.Role, &user.IsVerified, &user.Status, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return nil, err
	}

	query = `SELECT id,label, address_line, address_line, pincode, city, state, country, created_at, updated_at
	FROM user_addresses WHERE user_id = $1`

	rows, err := repo.DB.Query(ctx, query, user.ID)
	if err != nil {
		return nil, err
	}
	var addr = &domain.UserAddress{}
	for rows.Next() {
		err = rows.Scan(&addr.ID, &addr.Label, &addr.AddressLine, &addr.AddressLine2, &addr.Pincode, &addr.City,
			&addr.State, &addr.Country, &addr.CreatedAt, &addr.UpdatedAt)
		if err != nil {
			return nil, err
		}

		user.Addresses = append(user.Addresses, addr)
	}

	return user, nil
}

func (repo *AdminRepository) ListUsers(ctx context.Context, filter *domain.UserFilter) ([]*domain.UserProfile, error) {
	query := `
		SELECT id, email, COALESCE(phone,''), first_name, last_name, role, is_verified, status, created_at, updated_at
		FROM users
		WHERE 1=1
	`
	args := []any{}
	argIndex := 1

	if filter != nil {
		if filter.Status != "" {
			query += fmt.Sprintf(" AND status = $%d", argIndex)
			args = append(args, filter.Status)
			argIndex++
		}
		if filter.Role != "" {
			query += fmt.Sprintf(" AND role = $%d", argIndex)
			args = append(args, filter.Role)
			argIndex++
		}
		if filter.Search != "" {
			query += fmt.Sprintf(
				" AND (first_name ILIKE $%d OR last_name ILIKE $%d OR email ILIKE $%d)",
				argIndex, argIndex, argIndex,
			)
			args = append(args, "%"+filter.Search+"%")
			argIndex++
		}
	}

	query += " ORDER BY created_at DESC"

	if filter != nil && filter.Limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d", argIndex)
		args = append(args, filter.Limit)
		argIndex++
	}
	offset := (filter.Page - 1) * filter.Limit
	if filter != nil && offset > 0 {
		query += fmt.Sprintf(" OFFSET $%d", argIndex)
		args = append(args, offset)
		argIndex++
	}

	rows, err := repo.DB.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*domain.UserProfile
	for rows.Next() {
		var user domain.UserProfile
		if err = rows.Scan(
			&user.ID, &user.Email, &user.Phone,
			&user.FirstName, &user.LastName, &user.Role,
			&user.IsVerified, &user.Status, &user.CreatedAt, &user.UpdatedAt,
		); err != nil {
			return nil, err
		}
		users = append(users, &user)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	query = `SELECT COUNT(id) FROM USERS`
	err = repo.DB.QueryRow(ctx, query).Scan(&filter.Total)
	if err != nil {
		filter.Total = -1
	}

	return users, nil
}

func (repo *AdminRepository) UpdateUserStatus(ctx context.Context, req *domain.UserStatusUpdateRequest) error {
	_, err := repo.DB.Exec(ctx,
		`UPDATE users SET status = $1, updated_at = NOW() WHERE id = $2`,
		req.Status, req.UserID)
	if err != nil {
		return err
	}
	return nil
}

func (repo *AdminRepository) DeleteUser(ctx context.Context, id int64) error {
	cmdTag, err := repo.DB.Exec(ctx, `DELETE FROM users WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if cmdTag.RowsAffected() == 0 {
		return fmt.Errorf("User does not exist")
	}
	return nil
}
