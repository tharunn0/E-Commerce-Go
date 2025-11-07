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

func (repo *AdminRepository) GetUsersByName(ctx context.Context, name string) ([]domain.User, error) {
	rows, err := repo.DB.Query(ctx,
		`SELECT id, email, phone, password, first_name, last_name, role, is_verified, status, created_at, updated_at
		 FROM users WHERE (first_name ILIKE $1 OR last_name ILIKE $1) AND status = 'active'
		 ORDER BY created_at DESC`,
		"%"+name+"%")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []domain.User
	for rows.Next() {
		var user domain.User
		err := rows.Scan(&user.ID, &user.Email, &user.Phone, &user.Password, &user.FirstName,
			&user.LastName, &user.Role, &user.IsVerified, &user.Status, &user.CreatedAt, &user.UpdatedAt)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}
	return users, nil
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

	// 	type UserAddress struct {
	// 		ID           int64     `json:"id"`
	// 		UserID       int64     `json:"user_id"`
	// 		Label        string    `json:"label,omitempty"`
	// 		AddressLine  string    `json:"address_line"`
	// 		AddressLine2 string    `json:"address_line_2,omitempty"`
	// 		Pincode      string    `json:"pincode"`
	// 		City         string    `json:"city"`
	// 		State        string    `json:"state,omitempty"`
	// 		Country      string    `json:"country"`
	// 		CreatedAt    time.Time `json:"created_at"`
	// 		UpdatedAt    time.Time `json:"updated_at"`
	// }

	query = `SELECT id,label, address_line, address_line, pincode, city, state, country, created_at, updated_at
	 FROM user_addresses WHERE user_id = $1`

	for _, u := range users {
		var userAddr []*domain.UserAddress
		rows, err = repo.DB.Query(ctx, query, u.ID)
		if err != nil {
			return nil, err
		}

		for rows.Next() {
			var addr domain.UserAddress
			if err = rows.Scan(&addr.ID, &addr.Label, &addr.AddressLine,
				&addr.AddressLine2, &addr.Pincode, &addr.City, &addr.State, &addr.Country, &addr.CreatedAt, &addr.UpdatedAt,
			); err != nil {
				return nil, err
			}
			userAddr = append(userAddr, &addr)
		}
		u.Addresses = userAddr
	}

	return users, nil
}

func (repo *AdminRepository) CountUsers(ctx context.Context, filter *domain.UserFilter) (int64, error) {
	query := `SELECT COUNT(*) FROM users WHERE 1=1`
	args := []interface{}{}
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
			query += fmt.Sprintf(" AND (first_name ILIKE $%d OR last_name ILIKE $%d OR email ILIKE $%d)", argIndex, argIndex, argIndex)
			args = append(args, "%"+filter.Search+"%")
			argIndex++
		}
	}

	var count int64
	err := repo.DB.QueryRow(ctx, query, args...).Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}

func (repo *AdminRepository) UpdateUserStatus(ctx context.Context, req *domain.UserStatusUpdateRequest) error {
	_, err := repo.DB.Exec(ctx,
		`UPDATE users SET status = $1, updated_at = NOW() WHERE id = $2`,
		req.Status, req.UserID)
	if err != nil {
		fmt.Println("Error updating user status: ", err)
		return err
	}
	return nil
}
