package repository

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/tharunn0/E-Commerce-Go/internal/apperror"
	"github.com/tharunn0/E-Commerce-Go/internal/domain"
)

type UserRepository struct {
	DB *pgxpool.Pool
}

func NewUserRepository(db *pgxpool.Pool) *UserRepository {
	return &UserRepository{
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

func (repo *UserRepository) GoogleSignIn(ctx context.Context, req *domain.GoogleSignInRequest) (*domain.User, error) {
	user := &domain.User{}

	// check if user exists with google sub
	query := `SELECT id, email, COALESCE(phone,''), password, first_name, last_name, role,is_verified , status FROM users
    WHERE status = 'active' AND google_id = $1;`
	err := repo.DB.QueryRow(ctx, query, req.Sub).Scan(&user.ID, &user.Email, &user.Phone, &user.Password, &user.FirstName,
		&user.LastName, &user.Role, &user.IsVerified, &user.Status)

	if err == nil {
		return user, nil
	}
	// check if user exists with email
	query = `SELECT id, email, COALESCE(phone,''), password, first_name, last_name, role,is_verified , status FROM users
    WHERE status = 'active' AND email = $1;`
	err = repo.DB.QueryRow(ctx, query, req.Email).Scan(&user.ID, &user.Email, &user.Phone, &user.Password, &user.FirstName,
		&user.LastName, &user.Role, &user.IsVerified, &user.Status)
	if err == nil {
		// update user with google sub
		cmdTag, err := repo.DB.Exec(ctx, `UPDATE users SET google_id = $1 WHERE id = $2`, req.Sub, user.ID)
		if err != nil {
			return nil, err
		}
		if cmdTag.RowsAffected() != 1 {
			return nil, errors.New("FAILED_TO_UPDATE_USER")
		}
		return user, nil
	}

	query = `INSERT INTO users (email,first_name, last_name, password,provider, is_verified, google_id)
    VALUES ($1,$2,$3,$4,$5,$6,$7) RETURNING id, email,first_name,last_name,role,is_verified,status;`

	err = repo.DB.QueryRow(ctx, query, req.Email, req.FirstName, req.LastName, "", "google", req.Verified, req.Sub).Scan(
		&user.ID,
		&user.Email,
		&user.FirstName,
		&user.LastName,
		&user.Role,
		&user.IsVerified,
		&user.Status)

	if err != nil {
		return nil, err
	}
	return user, nil
}

func (repo *UserRepository) GetUserByID(ctx context.Context, userID int64) (*domain.User, error) {
	user := &domain.User{}
	query := `SELECT id, email, phone, first_name, last_name, role,is_verified , status, created_at, updated_at FROM users
    WHERE status = 'active' AND id = $1;`
	err := repo.DB.QueryRow(ctx, query, userID).Scan(&user.ID, &user.Email, &user.Phone, &user.FirstName,
		&user.LastName, &user.Role, &user.IsVerified, &user.Status, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (repo *UserRepository) UpdateUserProfile(ctx context.Context, userID int64, req *domain.UpdateUserProfileRequest) (*domain.UserProfile, error) {
	query := `
	UPDATE users SET
        first_name = COALESCE($1, first_name),
        last_name  = COALESCE($2, last_name),
        phone      = COALESCE($3, phone)
    WHERE id = $4
    RETURNING id, email, first_name, last_name, phone, is_verified, created_at`

	var updatedProfile domain.UserProfile
	err := repo.DB.QueryRow(ctx, query, req.FirstName, req.LastName, req.Phone, userID).Scan(
		&updatedProfile.ID, &updatedProfile.Email, &updatedProfile.FirstName, &updatedProfile.LastName,
		&updatedProfile.Phone, &updatedProfile.IsVerified, &updatedProfile.CreatedAt)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			switch pgErr.ConstraintName {
			case "users_phone_key":
				return nil, apperror.ErrPhoneExists
			}
		}
		return nil, err
	}

	return &updatedProfile, nil
}

// user address repository
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

func (repo *UserRepository) GetDefaultUserAddress(ctx context.Context, userID int64) (*domain.UserAddress, error) {

	userAddress := &domain.UserAddress{}
	query := `SELECT ua.id, ua.user_id, ua.label, ua.address_line,ua.address_line_2, ua.pincode, ua.city,
	 ua.state, ua.country , ua.created_at, ua.updated_at FROM user_addresses ua
	 INNER JOIN users u ON ua.user_id = u.id
	 WHERE ua.user_id = $1 AND u.default_address_id = ua.id`
	err := repo.DB.QueryRow(ctx, query, userID).Scan(&userAddress.ID, &userAddress.UserID, &userAddress.Label, &userAddress.AddressLine, &userAddress.AddressLine2,
		&userAddress.Pincode, &userAddress.City, &userAddress.State, &userAddress.Country, &userAddress.CreatedAt, &userAddress.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return userAddress, nil
}

func (repo *UserRepository) UpdateDefaultUserAddress(ctx context.Context, userID int64, addressID int64) error {
	query := `UPDATE users 
		SET default_address_id = $1 
			WHERE id = $2 
			AND EXISTS (
			SELECT 1 FROM user_addresses 
      WHERE id = $1 AND user_id = $2 
  )
`
	cmdTag, err := repo.DB.Exec(ctx, query, addressID, userID)
	if err != nil {
		return err
	}
	if cmdTag.RowsAffected() != 1 {
		return errors.New("FAILED_TO_UPDATE_DEFAULT_USER_ADDRESS")
	}
	return nil
}

func (repo *UserRepository) UpdateUserAddress(ctx context.Context, userID int64, req *domain.UpdateUserAddressRequest) (*domain.UserAddress, error) {

	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM user_addresses WHERE user_id = $1 AND id = $2)`
	err := repo.DB.QueryRow(ctx, query, userID, req.ID).Scan(&exists)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, apperror.ErrAddressNotFoundForUser
	}
	query = `UPDATE user_addresses SET
		label = COALESCE($1, label),
		address_line = COALESCE($2, address_line),
		address_line_2 = COALESCE($3, address_line_2),
		pincode = COALESCE($4, pincode),
		city = COALESCE($5, city),
		state = COALESCE($6, state),
		country = COALESCE($7, country)
	WHERE id = $8 
	RETURNING id, label, address_line, address_line_2, pincode, city, state, country
	`
	var updatedAddress domain.UserAddress
	err = repo.DB.QueryRow(ctx, query, req.Label, req.AddressLine, req.AddressLine2, req.Pincode, req.City, req.State, req.Country, req.ID).Scan(
		&updatedAddress.ID, &updatedAddress.Label, &updatedAddress.AddressLine, &updatedAddress.AddressLine2, &updatedAddress.Pincode, &updatedAddress.City, &updatedAddress.State, &updatedAddress.Country)
	if err != nil {
		return nil, err
	}
	return &updatedAddress, nil
}

func (repo *UserRepository) DeleteUserAddress(ctx context.Context, userID int64, addressID int64) error {

	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM user_addresses WHERE user_id = $1 AND id = $2)`
	err := repo.DB.QueryRow(ctx, query, userID, addressID).Scan(&exists)
	if err != nil {
		return err
	}
	if !exists {
		return apperror.ErrAddressNotFoundForUser
	}

	query = `DELETE FROM user_addresses WHERE user_id = $1 AND id = $2`
	cmdTag, err := repo.DB.Exec(ctx, query, userID, addressID)
	if err != nil {
		return err
	}
	if cmdTag.RowsAffected() != 1 {
		return errors.New("FAILED_TO_DELETE_USER_ADDRESS")
	}
	return nil
}
