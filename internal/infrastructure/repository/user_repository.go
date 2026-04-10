package repository

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/tharunn0/E-Commerce-Go/internal/apperror"
	"github.com/tharunn0/E-Commerce-Go/internal/domain/payment"
	"github.com/tharunn0/E-Commerce-Go/internal/domain/user"
)

type UserRepository struct {
	DB *pgxpool.Pool
}

func NewUserRepository(db *pgxpool.Pool) *UserRepository {
	return &UserRepository{
		DB: db,
	}
}

func (repo *UserRepository) RegisterUser(ctx context.Context, req *user.RegisterRequest, refCode string) error {

	tx, err := repo.DB.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	userID := int64(0)
	err = tx.QueryRow(ctx,
		`INSERT INTO users (first_name,last_name,email,phone,password, referral_code)
		 VALUES ($1,$2,$3,$4,$5,$6) RETURNING id`,
		req.FirstName,
		req.LastName,
		req.Email,
		req.Phone,
		req.Password,
		refCode,
	).Scan(&userID)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			switch pgErr.ConstraintName {
			case "users_email_key":
				return apperror.ErrEmailExists
			case "users_phone_key":
				return apperror.ErrPhoneExists
			}
		}
		return err
	}

	// create cart
	_, _ = tx.Exec(ctx,
		`INSERT INTO carts (user_id) VALUES ($1)`,
		userID,
	)

	// create wallet
	_, _ = tx.Exec(ctx,
		`INSERT INTO wallets (user_id,balance,is_admin)
		 VALUES ($1,0,$2)`,
		userID,
		false,
	)

	// create wishlist
	_, _ = tx.Exec(ctx,
		`INSERT INTO wishlists (user_id) VALUES ($1)`,
		userID,
	)

	// ===== referral handling (fixed logic) =====
	if req.ReferralCode != "" {

		var referrerUserID int64
		err = tx.QueryRow(ctx,
			`SELECT id FROM users WHERE referral_code = $1`,
			req.ReferralCode,
		).Scan(&referrerUserID)

		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return apperror.ErrInvalidReferralCode
			}
			return err
		}

		// prevent self-referral
		if referrerUserID == userID {
			return apperror.ErrInvalidReferralCode
		}

		// insert into referrals
		_, _ = tx.Exec(ctx,
			`INSERT INTO referrals (referrer_user_id,referred_user_id,referral_code,status)
			 VALUES ($1,$2,$3,$4)`,
			referrerUserID,
			userID,
			req.ReferralCode,
			"pending",
		)
	}

	// commit transaction
	if err = tx.Commit(ctx); err != nil {
		return err
	}

	return nil
}

func (repo *UserRepository) GetUser(ctx context.Context, email string) (*user.User, error) {
	user := &user.User{}
	query := `SELECT id, email, phone, password, first_name, last_name, role,is_verified , status, profile_img_url FROM users
    WHERE status = 'active' AND email = $1;`
	err := repo.DB.QueryRow(ctx,
		query, email).Scan(&user.ID, &user.Email, &user.Phone, &user.Password, &user.FirstName,
		&user.LastName, &user.Role, &user.IsVerified, &user.Status, &user.ProfilePicture)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperror.ErrUserNotFound
		}
		return nil, err
	}
	return user, nil
}

func (repo *UserRepository) GoogleSignIn(ctx context.Context, req *user.GoogleSignInRequest) (*user.User, error) {
	user := &user.User{}

	// check if user exists with google sub
	query := `SELECT id, email, COALESCE(phone,''), password, first_name, last_name, role,is_verified , status, profile_img_url FROM users
    WHERE status = 'active' AND google_id = $1;`
	err := repo.DB.QueryRow(ctx, query, req.Sub).Scan(&user.ID, &user.Email, &user.Phone, &user.Password, &user.FirstName,
		&user.LastName, &user.Role, &user.IsVerified, &user.Status, &user.ProfilePicture)

	if err == nil {
		return user, nil
	}
	// check if user exists with email
	query = `SELECT id, email, COALESCE(phone,''), password, first_name, last_name, role,is_verified , status, profile_img_url FROM users
    WHERE status = 'active' AND email = $1;`
	err = repo.DB.QueryRow(ctx, query, req.Email).Scan(&user.ID, &user.Email, &user.Phone, &user.Password, &user.FirstName,
		&user.LastName, &user.Role, &user.IsVerified, &user.Status, &user.ProfilePicture)
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

	query = `INSERT INTO users (email,first_name, last_name, password,provider, is_verified, google_id, profile_img_url)
    VALUES ($1,$2,$3,$4,$5,$6,$7,$8) RETURNING id, email,first_name,last_name,role,is_verified,status;`

	err = repo.DB.QueryRow(ctx, query, req.Email, req.FirstName, req.LastName, "", "google", req.Verified, req.Sub, "").Scan(
		&user.ID,
		&user.Email,
		&user.FirstName,
		&user.LastName,
		&user.Role,
		&user.IsVerified,
		&user.Status,
		&user.ProfilePicture)

	if err != nil {
		return nil, err
	}
	return user, nil
}

func (repo *UserRepository) GetUserByID(ctx context.Context, userID int64) (*user.User, error) {
	user := &user.User{}
	query := `SELECT id, email, phone, first_name, last_name, role,is_verified , status,COALESCE(referral_code, ""),profile_img_url,created_at, updated_at FROM users
    WHERE status = 'active' AND id = $1;`
	err := repo.DB.QueryRow(ctx, query, userID).Scan(&user.ID, &user.Email, &user.Phone, &user.FirstName,
		&user.LastName, &user.Role, &user.IsVerified, &user.Status, &user.ReferralCode, &user.ProfilePicture, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (repo *UserRepository) UpdateUserProfile(ctx context.Context, userID int64, req *user.UpdateUserProfileRequest) (*user.UserProfile, error) {
	query := `
	UPDATE users SET
        first_name = COALESCE($1, first_name),
        last_name  = COALESCE($2, last_name),
        phone      = COALESCE($3, phone),
		profile_img_url = COALESCE($4, profile_img_url)
    WHERE id = $5
    RETURNING id, email, first_name, last_name, phone, is_verified, created_at, profile_img_url`

	var updatedProfile user.UserProfile
	err := repo.DB.QueryRow(ctx, query, req.FirstName, req.LastName, req.Phone, req.ProfilePicture, userID).Scan(
		&updatedProfile.ID, &updatedProfile.Email, &updatedProfile.FirstName, &updatedProfile.LastName,
		&updatedProfile.Phone, &updatedProfile.IsVerified, &updatedProfile.CreatedAt, &updatedProfile.ProfilePicture)

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
func (repo *UserRepository) InsertUserAddress(ctx context.Context, address *user.UserAddress) error {
	query := `INSERT INTO user_addresses (user_id, label, address_line,address_line_2, pincode, city, district, state, country)
	 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)`
	cmdTag, err := repo.DB.Exec(ctx, query, address.UserID, address.Label, address.AddressLine, address.AddressLine2, address.Pincode, address.City, address.District, address.State, address.Country)
	if err != nil {
		return err
	}
	if cmdTag.RowsAffected() != 1 {
		return errors.New("FAILED_TO_INSERT_USER_ADDRESS")
	}
	return nil
}

func (repo *UserRepository) GetUserAddresses(ctx context.Context, userID int64) ([]*user.UserAddress, error) {
	query := `SELECT id, user_id, label, address_line,address_line_2, pincode, city, district, state, country , created_at, updated_at FROM user_addresses
	 WHERE user_id = $1`
	rows, err := repo.DB.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	userAddresses := []*user.UserAddress{}
	for rows.Next() {
		var userAddress user.UserAddress
		err := rows.Scan(&userAddress.ID, &userAddress.UserID, &userAddress.Label, &userAddress.AddressLine,
			&userAddress.AddressLine2, &userAddress.Pincode, &userAddress.City, &userAddress.District,
			&userAddress.State, &userAddress.Country, &userAddress.CreatedAt, &userAddress.UpdatedAt)
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

func (repo *UserRepository) GetUserAddressByID(ctx context.Context, addressID int64) (*user.UserAddress, error) {
	userAddress := &user.UserAddress{}
	query := `SELECT id, user_id, label, address_line,address_line_2, pincode, city, district, state, country , created_at, updated_at FROM user_addresses
	 WHERE id = $1`
	err := repo.DB.QueryRow(ctx, query, addressID).Scan(&userAddress.ID, &userAddress.UserID, &userAddress.Label, &userAddress.AddressLine,
		&userAddress.AddressLine2, &userAddress.Pincode, &userAddress.City, &userAddress.District,
		&userAddress.State, &userAddress.Country, &userAddress.CreatedAt, &userAddress.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperror.ErrAddressNotFoundForUser
		}
		return nil, err
	}
	return userAddress, nil
}

func (repo *UserRepository) GetDefaultUserAddress(ctx context.Context, userID int64) (*user.UserAddress, error) {

	userAddress := &user.UserAddress{}
	query := `SELECT ua.id, ua.user_id, ua.label, ua.address_line,ua.address_line_2, ua.pincode, ua.city,
	 ua.district, ua.state, ua.country , ua.created_at, ua.updated_at FROM user_addresses ua
	 INNER JOIN users u ON ua.user_id = u.id
	 WHERE ua.user_id = $1 AND u.default_address_id = ua.id`
	err := repo.DB.QueryRow(ctx, query, userID).Scan(&userAddress.ID, &userAddress.UserID, &userAddress.Label, &userAddress.AddressLine, &userAddress.AddressLine2,
		&userAddress.Pincode, &userAddress.City, &userAddress.District, &userAddress.State, &userAddress.Country, &userAddress.CreatedAt, &userAddress.UpdatedAt)
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

func (repo *UserRepository) UpdateUserAddress(ctx context.Context, userID int64, req *user.UpdateUserAddressRequest) (*user.UserAddress, error) {

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
		district = COALESCE($6, district),
		state = COALESCE($7, state),
		country = COALESCE($8, country)
	WHERE id = $9 
	RETURNING id, label, address_line, address_line_2, pincode, city, district, state, country
	`
	var updatedAddress user.UserAddress
	err = repo.DB.QueryRow(ctx, query, req.Label, req.AddressLine, req.AddressLine2, req.Pincode, req.City, req.District, req.State, req.Country, req.ID).Scan(
		&updatedAddress.ID, &updatedAddress.Label, &updatedAddress.AddressLine, &updatedAddress.AddressLine2, &updatedAddress.Pincode, &updatedAddress.City, &updatedAddress.District, &updatedAddress.State, &updatedAddress.Country)
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

// user wallet ops

func (repo *UserRepository) GetWallet(ctx context.Context, userID int64) (*payment.Wallet, error) {
	wallet := &payment.Wallet{}
	query := `SELECT id, balance, created_at, updated_at FROM wallets
    WHERE user_id = $1;`
	err := repo.DB.QueryRow(ctx, query, userID).Scan(&wallet.ID, &wallet.Balance, &wallet.CreatedAt, &wallet.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return wallet, nil
}

func (repo *UserRepository) GetWalletTransactions(ctx context.Context, userID int64, filter *payment.TransactionFilter) ([]*payment.WalletTransaction, error) {

	query := `
	SELECT
		wt.id,
		wt.wallet_id,
		wt.amount,
		wt.transaction_type,
		wt.related_order,
		wt.remarks,
		wt.balance_before,
		wt.balance_after,
		wt.created_at
	FROM wallet_transactions wt
	JOIN wallets w ON w.id = wt.wallet_id
	WHERE w.user_id = $1
	`

	args := []interface{}{userID}
	paramIndex := 2

	if filter != nil {

		// if type is empty skip it, else add it to the query
		if filter.Type != "" {
			query += fmt.Sprintf(" AND wt.transaction_type = $%d", paramIndex)
			args = append(args, filter.Type)
			paramIndex++
		}

		query += fmt.Sprintf(" AND wt.created_at BETWEEN $%d AND $%d", paramIndex, paramIndex+1)
		args = append(args, filter.StartDate, filter.EndDate)
		paramIndex += 2
	}

	// Always sort transactions
	query += " ORDER BY wt.created_at DESC"

	if filter != nil && filter.Limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d", paramIndex)
		args = append(args, filter.Limit)
		paramIndex++
	}

	if filter != nil && filter.Page > 0 && filter.Limit > 0 {
		offset := (filter.Page - 1) * filter.Limit
		query += fmt.Sprintf(" OFFSET $%d", paramIndex)
		args = append(args, offset)
	}

	log.Println("query", query)

	rows, err := repo.DB.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var transactions []*payment.WalletTransaction

	for rows.Next() {
		t := &payment.WalletTransaction{}

		err := rows.Scan(
			&t.ID,
			&t.WalletID,
			&t.Amount,
			&t.Type,
			&t.RelatedOrderID,
			&t.Remark,
			&t.BalanceBefore,
			&t.BalanceAfter,
			&t.CreatedAt,
		)
		if err != nil {
			return nil, err
		}

		transactions = append(transactions, t)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return transactions, nil
}
