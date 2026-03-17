package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/tharunn0/E-Commerce-Go/internal/apperror"
	"github.com/tharunn0/E-Commerce-Go/internal/domain/user"
)

type AdminRepository struct {
	DB *pgxpool.Pool
}

func NewAdminRepository(db *pgxpool.Pool) *AdminRepository {
	return &AdminRepository{
		DB: db,
	}
}

func (repo *AdminRepository) CreateAdmin(ctx context.Context, req *user.AdminRegisterRequest) error {
	query := `INSERT INTO users (email, phone, password, first_name, last_name, role,is_verified) VALUES ($1, $2, $3, $4, $5, $6, $7)`
	cmdTag, err := repo.DB.Exec(ctx, query, req.Email, req.Phone, req.Password, req.FirstName, req.LastName, "admin", true)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == "23505" {
				switch pgErr.ConstraintName {
				case "users_email_key":
					return apperror.ErrEmailExists
				case "users_phone_key":
					return apperror.ErrPhoneExists
				}
			}
		}
		return err
	}
	if cmdTag.RowsAffected() != 1 {
		return fmt.Errorf("FAILED_TO_CREATE_ADMIN")
	}

	return nil
}

func (repo *AdminRepository) GetAdmin(ctx context.Context, email string) (*user.User, error) {
	userObj := &user.User{}
	err := repo.DB.QueryRow(ctx,
		`SELECT id, email, phone, password, first_name, last_name, role,is_verified , status FROM users
    WHERE role = 'admin' AND email = $1;
	`, email).Scan(&userObj.ID, &userObj.Email, &userObj.Phone, &userObj.Password, &userObj.FirstName,
		&userObj.LastName, &userObj.Role, &userObj.IsVerified, &userObj.Status)
	if err != nil {
		return nil, err
	}
	return userObj, nil
}

func (repo *AdminRepository) GetUserByEmail(ctx context.Context, email string) (*user.User, error) {
	userObj := &user.User{}
	err := repo.DB.QueryRow(ctx,
		`SELECT id, email, phone, password, first_name, last_name, role, is_verified, status, created_at, updated_at
		 FROM users WHERE email = $1 AND status = 'active'`,
		email).Scan(&userObj.ID, &userObj.Email, &userObj.Phone, &userObj.Password, &userObj.FirstName,
		&userObj.LastName, &userObj.Role, &userObj.IsVerified, &userObj.Status, &userObj.CreatedAt, &userObj.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return userObj, nil
}

func (repo *AdminRepository) GetUsersByID(ctx context.Context, id int64) (*user.UserProfile, error) {

	userProfile := &user.UserProfile{}

	query := `SELECT id, email,COALESCE(phone,''), first_name, last_name, role, is_verified, status, created_at, updated_at FROM users
	WHERE id = $1`

	err := repo.DB.QueryRow(ctx, query, id).Scan(&userProfile.ID, &userProfile.Email, &userProfile.Phone, &userProfile.FirstName,
		&userProfile.LastName, &userProfile.Role, &userProfile.IsVerified, &userProfile.Status, &userProfile.CreatedAt, &userProfile.UpdatedAt)
	if err != nil {
		return nil, err
	}

	query = `SELECT id,label, address_line, address_line, pincode, city, state, country, created_at, updated_at
	FROM user_addresses WHERE user_id = $1`

	rows, err := repo.DB.Query(ctx, query, userProfile.ID)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		addr := &user.UserAddress{}
		err = rows.Scan(&addr.ID, &addr.Label, &addr.AddressLine, &addr.AddressLine2, &addr.Pincode, &addr.City,
			&addr.State, &addr.Country, &addr.CreatedAt, &addr.UpdatedAt)
		if err != nil {
			return nil, err
		}

		userProfile.Addresses = append(userProfile.Addresses, addr)
	}

	return userProfile, nil
}

func (repo *AdminRepository) ListUsers(ctx context.Context, filter *user.UserFilter) ([]*user.UserProfile, error) {
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

	var users []*user.UserProfile
	for rows.Next() {
		var userObj user.UserProfile
		if err = rows.Scan(
			&userObj.ID, &userObj.Email, &userObj.Phone,
			&userObj.FirstName, &userObj.LastName, &userObj.Role,
			&userObj.IsVerified, &userObj.Status, &userObj.CreatedAt, &userObj.UpdatedAt,
		); err != nil {
			return nil, err
		}
		users = append(users, &userObj)
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

func (repo *AdminRepository) UpdateUserStatus(ctx context.Context, req *user.UserStatusUpdateRequest) error {
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

