package service

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/tharunn0/E-Commerce-Go/internal/apperror"
	"github.com/tharunn0/E-Commerce-Go/internal/domain"
	"github.com/tharunn0/E-Commerce-Go/internal/utils"
	"github.com/tharunn0/E-Commerce-Go/pkg/mailer"

	"go.uber.org/zap"
)

type UserService struct {
	repo   domain.UserRepository
	auth   domain.AuthRepository
	sender *mailer.MailSender
	log    *zap.Logger
}

func NewUserService(userRepo domain.UserRepository, authrepo domain.AuthRepository, logger *zap.Logger, sender *mailer.MailSender) *UserService {
	return &UserService{
		repo:   userRepo,
		log:    logger,
		sender: sender,
		auth:   authrepo,
	}
}

// register a new user
func (serv *UserService) RegisterUser(ctx context.Context, req *domain.RegisterRequest) *apperror.APIError {
	serv.log.Debug("starting user registration", zap.String("email", req.Email))

	if !utils.IsValidEmail(req.Email) {
		serv.log.Warn("invalid email format", zap.String("email", req.Email))
		return &apperror.APIError{
			Status:  http.StatusBadRequest,
			Code:    "INVALID_EMAIL",
			Message: "Please provide a valid email.",
		}
	}

	if !utils.IsValidPassword(req.Password) {
		serv.log.Warn("weak password", zap.String("email", req.Email))
		return &apperror.APIError{
			Status:  http.StatusBadRequest,
			Code:    "WEAK_PASSWORD",
			Message: "Password must contain at least 8 characters, including numbers/symbols.",
		}
	}

	if req.Password != req.ConfirmPassword {
		serv.log.Warn("passwords do not match", zap.String("email", req.Email))
		return &apperror.APIError{
			Status:  http.StatusBadRequest,
			Code:    "PASSWORDS_DONT_MATCH",
			Message: "Passwords do not match.",
		}
	}

	fmt.Println("Used Referral Code: ", req.ReferralCode)

	existingUser, err := serv.repo.GetUser(ctx, req.Email)
	if err == nil && existingUser != nil {
		serv.log.Warn("duplicate user registration attempt", zap.String("email", req.Email))
		return &apperror.APIError{
			Status:  http.StatusConflict,
			Code:    "EMAIL_ALREADY_EXISTS",
			Message: "An account with this email already exists.",
		}
	}

	hashedPass := utils.HashPassword(req.Password)
	if len(hashedPass) == 0 {
		serv.log.Error("password hashing failed", zap.Error(err))
		return &apperror.APIError{
			Status:  http.StatusInternalServerError,
			Code:    "HASHING_FAILED",
			Message: "Could not process password. Please try again.",
		}
	}

	req.Password = hashedPass
	// serv.log.Debug("registering user", zap.String("service", "UserService"), zap.Any("request", req))

	// generate referral code

	refCode, err := utils.GenerateReferralCode(req.FirstName)
	if err != nil {
		serv.log.Error("failed to generate referral code", zap.String("email", req.Email))
		return &apperror.APIError{
			Status:  http.StatusInternalServerError,
			Code:    "REFERRAL_CODE_GENERATION_FAILED",
			Message: "Could not generate referral code. Please try again.",
		}
	}

	err = serv.repo.RegisterUser(ctx, req, refCode)
	if err != nil {
		serv.log.Error("user registration failed", zap.String("email", req.Email), zap.Error(err))
		return &apperror.APIError{
			Status:  http.StatusInternalServerError,
			Code:    "DB_ERROR",
			Message: err.Error(),
		}
	}

	serv.log.Debug("user registered successfully", zap.String("email", req.Email))
	return nil
}

// login user
func (serv *UserService) LoginUser(ctx context.Context, req *domain.LoginRequest) (*domain.LoginResponse, *apperror.APIError) {

	if !utils.IsValidEmail(req.Email) {
		serv.log.Warn("invalid email format", zap.String("email", req.Email))
		return nil, &apperror.APIError{
			Status:  http.StatusBadRequest,
			Code:    "INVALID_EMAIL",
			Message: "Please provide a valid email address.",
		}
	}

	fetchedUser, err := serv.repo.GetUser(ctx, req.Email)
	if err != nil {
		if err == apperror.ErrUserNotFound {
			serv.log.Warn("login failed: user not found", zap.String("email", req.Email))
			return nil, &apperror.APIError{
				Status:  http.StatusNotFound,
				Code:    "USER_NOT_FOUND",
				Message: "No account found with this email.",
			}
		}

		serv.log.Error("failed to fetch user from db",
			zap.String("email", req.Email),
			zap.Error(err),
		)
		return nil, &apperror.APIError{
			Status:  http.StatusInternalServerError,
			Code:    "DB_ERROR",
			Message: "Something went wrong. Please try again later.",
		}
	}

	if fetchedUser.Status == "blocked" || fetchedUser.Status == "deleted" {
		serv.log.Warn("blocked or deleted user attempted login", zap.Int64("user_id", fetchedUser.ID))
		return nil, &apperror.APIError{
			Status:  http.StatusForbidden,
			Code:    "USER_BLOCKED",
			Message: "Your account is not active. Please contact support.",
		}
	}

	if !utils.VerifyPassword(fetchedUser.Password, req.Password) {
		serv.log.Warn("wrong password", zap.String("email", req.Email))
		return nil, &apperror.APIError{
			Status:  http.StatusUnauthorized,
			Code:    "WRONG_PASSWORD",
			Message: "Invalid email or password.",
		}
	}
	resp := domain.LoginResponse{}
	resp.User.ID = fetchedUser.ID
	resp.User.Email = fetchedUser.Email
	resp.User.FirstName = fetchedUser.FirstName
	resp.User.LastName = fetchedUser.LastName
	resp.User.Role = string(fetchedUser.Role)

	resp.AccessToken, err = utils.IssueJWT(fetchedUser.ID, fetchedUser.Email, fetchedUser.Role, fetchedUser.IsVerified, serv.log)
	if err != nil {
		serv.log.Error("Failed to issue jwt", zap.String("service", "UserService"), zap.Error(err))
		return nil, &apperror.APIError{
			Status:  http.StatusInternalServerError,
			Code:    "JWT_GENERATION_FAILED",
			Message: "Could not generate token. Please try again.",
		}
	}
	var expiryAt time.Time
	resp.RefreshToken, expiryAt, err = utils.GenerateTokenWithExpiry(32, 15*1440)
	if err != nil {
		serv.log.Error("Failed to generate refresh token", zap.String("service", "UserService"), zap.Error(err))
		return nil, &apperror.APIError{
			Status:  http.StatusInternalServerError,
			Code:    "REFRESH_TOKEN_GENERATION_FAILED",
			Message: "Could not generate refresh token. Please try again.",
		}
	}

	err = serv.auth.SetRefreshToken(ctx, &domain.RefreshToken{
		UserID:   fetchedUser.ID,
		Token:    resp.RefreshToken,
		ExpiryAt: expiryAt,
		Revoked:  false,
	})
	if err != nil {
		serv.log.Error("Failed to set refresh token", zap.String("service", "UserService"), zap.Error(err))
		return nil, &apperror.APIError{
			Status:  http.StatusInternalServerError,
			Code:    "REFRESH_TOKEN_SET_FAILED",
			Message: "Failed to set refresh token. Please try again.",
		}
	}
	return &resp, nil
}

// oauth sign in
func (serv *UserService) OAuthSignIn(ctx context.Context, req *domain.GoogleSignInRequest) (*domain.LoginResponse, *apperror.APIError) {

	user, err := serv.repo.GoogleSignIn(ctx, req)
	if user == nil {
		return nil, &apperror.APIError{
			Status:  http.StatusNotFound,
			Code:    "USER_NOT_FOUND",
			Message: "No account found with this email.",
		}
	}
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, &apperror.APIError{
				Status:  http.StatusNotFound,
				Code:    "USER_NOT_FOUND",
				Message: err.Error(),
			}
		}
		return nil, &apperror.APIError{
			Status:  http.StatusInternalServerError,
			Code:    "DB_ERROR",
			Message: err.Error(),
		}
	}

	authTokens := domain.AuthTokens{}
	authTokens.AccessToken, err = utils.IssueJWT(user.ID, user.Email, user.Role, user.IsVerified, serv.log)
	if err != nil {
		serv.log.Error("Failed to issue jwt", zap.String("service", "UserService"), zap.Error(err))
		return nil, &apperror.APIError{
			Status:  http.StatusInternalServerError,
			Code:    "JWT_GENERATION_FAILED",
			Message: "Could not generate token. Please try again.",
		}
	}
	authTokens.RefreshToken, err = utils.GenerateToken(32)
	if err != nil {
		serv.log.Error("Failed to generate refresh token", zap.String("service", "UserService"), zap.Error(err))
		return nil, &apperror.APIError{
			Status:  http.StatusInternalServerError,
			Code:    "REFRESH_TOKEN_GENERATION_FAILED",
			Message: "Could not generate refresh token. Please try again.",
		}
	}

	resp := domain.LoginResponse{
		AccessToken:  authTokens.AccessToken,
		RefreshToken: authTokens.RefreshToken,
	}

	resp.User.ID = user.ID
	resp.User.Email = user.Email
	resp.User.FirstName = user.FirstName
	resp.User.LastName = user.LastName
	resp.User.Role = string(user.Role)

	return &resp, nil
}

// get user profile
func (serv *UserService) GetUserProfile(ctx context.Context) (*domain.UserProfile, *apperror.APIError) {

	userID, err := utils.GetUserIDFromContext(ctx)
	if err != nil {
		return nil, &apperror.APIError{
			Status:  http.StatusBadRequest,
			Code:    "INVALID_REQUEST",
			Message: "User ID not found",
		}
	}

	isAdmin := utils.IsAdmin(ctx)

	user, err := serv.repo.GetUserByID(ctx, userID)
	if err != nil {
		serv.log.Error("failed to fetch user profile from db", zap.Int64("user_id", userID), zap.Error(err))
		return nil, &apperror.APIError{
			Status:  http.StatusNotFound,
			Code:    "DB_ERROR",
			Message: "Failed to fetch user profile.",
		}
	}

	addresses, err := serv.repo.GetUserAddresses(ctx, userID)
	if err != nil {
		serv.log.Error("failed to fetch user addresses from db", zap.Int64("user_id", userID), zap.Error(err))
		return nil, &apperror.APIError{
			Status:  http.StatusNotFound,
			Code:    "DB_ERROR",
			Message: "Failed to fetch user addresses.",
		}
	}

	var userProfile = domain.UserProfile{
		ID:               user.ID,
		Email:            user.Email,
		FirstName:        user.FirstName,
		LastName:         user.LastName,
		Phone:            &user.Phone,
		Role:             &user.Role,
		IsVerified:       user.IsVerified,
		Status:           &user.Status,
		ProfilePicture:   user.ProfilePicture,
		CreatedAt:        &user.CreatedAt,
		DefaultAddressID: user.DefaultAddressID,
		ReferralCode:     user.ReferralCode,
		Addresses:        addresses,
	}

	if user.DefaultAddressID == nil {
		if len(addresses) > 0 {
			addressID := addresses[0].ID
			userProfile.DefaultAddressID = &addressID
		}

	}

	if !isAdmin {
		userProfile.Status = nil
		userProfile.Role = nil

		for _, address := range userProfile.Addresses {
			address.UpdatedAt = nil
		}
	}

	return &userProfile, nil
}

// update user profile
func (serv *UserService) UpdateUserProfile(ctx context.Context, req *domain.UpdateUserProfileRequest) (*domain.UserProfile, *apperror.APIError) {

	err := req.Validate()
	if err != nil {
		return nil, &apperror.APIError{
			Status:  http.StatusBadRequest,
			Code:    "INVALID_REQUEST",
			Message: err.Error(),
		}
	}

	userID, err := utils.GetUserIDFromContext(ctx)
	if err != nil {
		return nil, &apperror.APIError{
			Status:  http.StatusUnauthorized,
			Code:    "INVALID_USER_ID",
			Message: "Invalid user ID.",
		}
	}
	serv.log.Debug("update user profile", zap.Int64("user id", userID))
	updatedProfile, err := serv.repo.UpdateUserProfile(ctx, userID, req)
	if err != nil {
		if err == apperror.ErrPhoneExists {
			return nil, &apperror.APIError{
				Status:  http.StatusConflict,
				Code:    "PHONE_ALREADY_EXISTS",
				Message: "A user with this phone number already exists.",
			}
		}
		serv.log.Debug("failed to update user profile in db", zap.Int64("user_id", userID), zap.Error(err))
		return nil, &apperror.APIError{
			Status:  http.StatusInternalServerError,
			Code:    "DB_ERROR",
			Message: "Failed to update user profile.",
		}
	}
	return updatedProfile, nil
}

// add user address
func (serv *UserService) CreateUserAddress(ctx context.Context, address *domain.UserAddress) *apperror.APIError {
	userID, err := utils.GetUserIDFromContext(ctx)
	if err != nil {
		return &apperror.APIError{
			Status:  http.StatusUnauthorized,
			Code:    "INVALID_USER_ID",
			Message: "Invalid user ID.",
		}
	}
	address.UserID = userID
	err = serv.repo.InsertUserAddress(ctx, address)
	if err != nil {
		serv.log.Debug("failed to insert user address into db", zap.Int64("user_id", address.UserID), zap.Error(err))
		return &apperror.APIError{
			Status:  http.StatusInternalServerError,
			Code:    "DB_ERROR",
			Message: "Failed to add user address.",
		}
	}
	return nil
}

// get user addresses
func (serv *UserService) GetUserAddresses(ctx context.Context) ([]*domain.UserAddress, *int64, *apperror.APIError) {

	isAdmin := utils.IsAdmin(ctx)

	userID, err := utils.GetUserIDFromContext(ctx)
	if err != nil {
		return nil, nil, &apperror.APIError{
			Status:  http.StatusUnauthorized,
			Code:    "INVALID_USER_ID",
			Message: "Invalid user ID.",
		}
	}

	addresses, err := serv.repo.GetUserAddresses(ctx, userID)
	if err != nil {
		serv.log.Debug("failed to fetch user addresses from db", zap.Int64("user_id", userID), zap.Error(err))
		return nil, nil, &apperror.APIError{
			Status:  http.StatusNotFound,
			Code:    "DB_ERROR",
			Message: "Failed to fetch user addresses.",
		}
	}

	if len(addresses) == 0 {
		return nil, nil, &apperror.APIError{
			Status:  http.StatusNotFound,
			Code:    "NOT_FOUND",
			Message: "User doesn't have any addresses.",
		}
	}

	fmt.Println("default address", addresses)

	defaultAddress, err := serv.repo.GetDefaultUserAddress(ctx, userID)
	if err != nil {
		serv.log.Debug("failed to fetch default user address from db", zap.Int64("user_id", userID), zap.Error(err))
		return nil, nil, &apperror.APIError{
			Status:  http.StatusNotFound,
			Code:    "NOT_FOUND",
			Message: "Default user address not found.",
		}
	}

	if !isAdmin {
		for _, address := range addresses {
			address.UpdatedAt = nil
		}
	}
	return addresses, &defaultAddress.ID, nil
}

// update default user address
func (serv *UserService) UpdateDefaultUserAddress(ctx context.Context, addressID int64) *apperror.APIError {
	userID, err := utils.GetUserIDFromContext(ctx)
	if err != nil {
		return &apperror.APIError{
			Status:  http.StatusUnauthorized,
			Code:    "INVALID_USER_ID",
			Message: "Invalid user ID.",
		}
	}
	err = serv.repo.UpdateDefaultUserAddress(ctx, userID, addressID)
	if err != nil {
		serv.log.Debug("failed to update default user address in db", zap.Int64("user_id", userID), zap.Int64("address_id", addressID), zap.Error(err))
		return &apperror.APIError{
			Status:  http.StatusConflict,
			Code:    "NOT_FOUND",
			Message: "Default user address not found.",
		}
	}
	return nil
}

// update user address
func (serv *UserService) UpdateUserAddress(ctx context.Context, req *domain.UpdateUserAddressRequest) (*domain.UserAddress, *apperror.APIError) {
	userID, err := utils.GetUserIDFromContext(ctx)
	if err != nil {
		return nil, &apperror.APIError{
			Status:  http.StatusUnauthorized,
			Code:    "INVALID_USER_ID",
			Message: "Invalid user ID.",
		}
	}

	updatedAddress, err := serv.repo.UpdateUserAddress(ctx, userID, req)
	if err != nil {
		serv.log.Debug("failed to update user address in db", zap.Int64("user_id", userID), zap.Int64("address_id", req.ID), zap.Error(err))

		if errors.Is(err, apperror.ErrAddressNotFoundForUser) {
			return nil, &apperror.APIError{
				Status:  http.StatusNotFound,
				Code:    "ADDRESS_NOT_FOUND",
				Message: "Address not found for user.",
			}
		}
		return nil, &apperror.APIError{
			Status:  http.StatusConflict,
			Code:    "DB_ERROR",
			Message: "Failed to update user address.",
		}
	}
	return updatedAddress, nil
}

// delete user address
func (serv *UserService) DeleteUserAddress(ctx context.Context, addressID int64) *apperror.APIError {
	userID, err := utils.GetUserIDFromContext(ctx)
	if err != nil {
		return &apperror.APIError{
			Status:  http.StatusUnauthorized,
			Code:    "INVALID_USER_ID",
			Message: "Invalid user ID.",
		}
	}
	err = serv.repo.DeleteUserAddress(ctx, userID, addressID)
	if err != nil {
		if errors.Is(err, apperror.ErrAddressNotFoundForUser) {
			return &apperror.APIError{
				Status:  http.StatusNotFound,
				Code:    "ADDRESS_NOT_FOUND",
				Message: "Address not found for user.",
			}
		}
		serv.log.Debug("failed to delete user address in db", zap.Int64("user_id", userID), zap.Int64("address_id", addressID), zap.Error(err))
		return &apperror.APIError{
			Status:  http.StatusInternalServerError,
			Code:    "DB_ERROR",
			Message: "Failed to delete user address.",
		}
	}
	return nil
}

func (serv *UserService) GetWallet(ctx context.Context) (*domain.Wallet, *apperror.APIError) {
	userID, err := utils.GetUserIDFromContext(ctx)
	if err != nil {
		return nil, &apperror.APIError{
			Status:  http.StatusUnauthorized,
			Code:    "INVALID_USER_ID",
			Message: "Invalid user ID.",
		}
	}
	wallet, err := serv.repo.GetWallet(ctx, userID)
	if err != nil {
		serv.log.Debug("failed to fetch wallet from db", zap.Int64("user_id", userID), zap.Error(err))
		return nil, &apperror.APIError{
			Status:  http.StatusNotFound,
			Code:    "NOT_FOUND",
			Message: "Wallet not found.",
		}
	}
	return wallet, nil
}

func (serv *UserService) GetWalletTransactions(ctx context.Context, filter *domain.TransactionFilter) ([]*domain.WalletTransaction, *apperror.APIError) {
	userID, err := utils.GetUserIDFromContext(ctx)
	if err != nil {
		return nil, &apperror.APIError{
			Status:  http.StatusUnauthorized,
			Code:    "INVALID_USER_ID",
			Message: "Invalid user ID.",
		}
	}

	if err := filter.Validate(); err != nil {
		return nil, &apperror.APIError{
			Status:  http.StatusBadRequest,
			Code:    "INVALID_REQUEST",
			Message: err.Error(),
		}
	}

	transactions, err := serv.repo.GetWalletTransactions(ctx, userID, filter)
	if err != nil {
		serv.log.Debug("failed to fetch wallet transactions from db", zap.Int64("user_id", userID), zap.Error(err))
		return nil, &apperror.APIError{
			Status:  http.StatusNotFound,
			Code:    "NOT_FOUND",
			Message: "Wallet transactions not found.",
		}
	}
	return transactions, nil
}
