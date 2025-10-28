package repository

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/tharunn0/E-Commerce-Go/internal/domain"
)

type UserRepository struct {
	DB *pgxpool.Pool
}

func NewUserRepository(db *pgxpool.Pool) UserRepository {
	return UserRepository{
		DB: db,
	}
}

func (repo *UserRepository) RegisterUser(ctx context.Context, req *domain.RegisterRequest) error {
	cmdTag, err := repo.DB.Exec(ctx,
		`INSERT INTO users (first_name,last_name,email,phone,password)
	 VALUES ($1,$2,$3,$4,$5)
	`,
		req.FirstName, req.LastName, req.Email, req.Phone, req.Password)
	if err != nil {
		return err
	}
	if cmdTag.RowsAffected() != 1 {
		return errors.New("USER_CREATION_FAILED")
	}
	return nil
}

func (repo *UserRepository) GetUser(ctx context.Context, email string) (*domain.User, error) {
	user := &domain.User{}
	query := `SELECT id, email, phone, password, first_name, last_name, role,is_verified , status FROM users
    WHERE status = 'active' AND email = $1;`
	err := repo.DB.QueryRow(ctx,
		query, email).Scan(&user.ID, &user.Email, &user.Phone, &user.Password, &user.FirstName,
		&user.LastName, &user.Role, &user.IsVerified, &user.Status)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (repo *UserRepository) GetUserByID(ctx context.Context, userID int64) (*domain.User, error) {
	user := &domain.User{}
	query := `SELECT id, email, phone, password, first_name, last_name, role,is_verified , status, created_at FROM users
    WHERE status = 'active' AND id = $1;`
	err := repo.DB.QueryRow(ctx, query, userID).Scan(&user.ID, &user.Email, &user.Phone, &user.Password, &user.FirstName,
		&user.LastName, &user.Role, &user.IsVerified, &user.Status, &user.CreatedAt)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (repo *UserRepository) InsertUserAddress(ctx context.Context, address *domain.UserAddress) error {
	query := `INSERT INTO user_addresses (user_id, label, address_line,address_line_2, pincode, city, state, country)
	 VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`
	cmdTag, err := repo.DB.Exec(ctx, query, address.UserID, address.Label, address.AddressLine, address.AddressLine2, address.Pincode, address.City, address.State, address.Country)
	if err != nil {
		return err
	}
	if cmdTag.RowsAffected() != 1 {
		return errors.New("FAILED_TO_INSERT_USER_ADDRESS")
	}
	return nil
}

func (repo *UserRepository) GetUserAddresses(ctx context.Context, userID int64) ([]*domain.UserAddress, error) {
	query := `SELECT id, user_id, label, address_line,address_line_2, pincode, city, state, country , created_at, updated_at FROM user_addresses
	 WHERE user_id = $1`
	rows, err := repo.DB.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	userAddresses := []*domain.UserAddress{}
	for rows.Next() {
		var userAddress domain.UserAddress
		err := rows.Scan(&userAddress.ID, &userAddress.UserID, &userAddress.Label, &userAddress.AddressLine,
			&userAddress.AddressLine2, &userAddress.Pincode, &userAddress.City, &userAddress.State,
			&userAddress.Country, &userAddress.CreatedAt, &userAddress.UpdatedAt)
		if err != nil {
			return nil, err
		}
		userAddresses = append(userAddresses, &userAddress)
	}
	if rows.Err() != nil {
		return nil, rows.Err()
	}
	return userAddresses, nil
}

func (repo *UserRepository) GetDefaultUserAddress(ctx context.Context, addressID int64) (*domain.UserAddress, error) {
	query := `SELECT id, user_id, label, address_line,address_line_2, pincode, city, state, country , created_at, updated_at FROM user_addresses
	 WHERE address_id = $1`
	rows, err := repo.DB.Query(ctx, query, addressID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var userAddress domain.UserAddress
	err = rows.Scan(&userAddress.ID, &userAddress.UserID, &userAddress.Label, &userAddress.AddressLine, &userAddress.AddressLine2, &userAddress.Pincode, &userAddress.City, &userAddress.State, &userAddress.Country, &userAddress.CreatedAt, &userAddress.UpdatedAt)
	if err != nil {
		return nil, err
	}
	if rows.Err() != nil {
		return nil, rows.Err()
	}
	return &userAddress, nil
}
