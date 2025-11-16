package service

import (
	"context"
	"errors"
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
			Code:    "INVALID_EMAIL",
			Message: "Please provide a valid email.",
		}
	}

	if !utils.IsValidPassword(req.Password) {
		serv.log.Warn("weak password", zap.String("email", req.Email))
		return &apperror.APIError{
			Code:    "WEAK_PASSWORD",
			Message: "Password must contain at least 8 characters, including numbers/symbols.",
		}
	}

	existingUser, err := serv.repo.GetUser(ctx, req.Email)
	if err == nil && existingUser != nil {
		serv.log.Warn("duplicate user registration attempt", zap.String("email", req.Email))
		return &apperror.APIError{
			Code:    "EMAIL_ALREADY_EXISTS",
			Message: "An account with this email already exists.",
		}
	}

	hashedPass := utils.HashPassword(req.Password)
	if len(hashedPass) == 0 {
		serv.log.Error("password hashing failed", zap.Error(err))
		return &apperror.APIError{
			Code:    "HASHING_FAILED",
			Message: "Could not process password. Please try again.",
		}
	}

	req.Password = hashedPass
	serv.log.Debug("registering user", zap.String("service", "UserService"), zap.Any("request", req))

	err = serv.repo.RegisterUser(ctx, req)
	if err != nil {
		serv.log.Error("user registration failed", zap.String("email", req.Email), zap.Error(err))
		return &apperror.APIError{
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
			Code:    "INVALID_EMAIL",
			Message: "Please provide a valid email address.",
		}
	}

	fetchedUser, err := serv.repo.GetUser(ctx, req.Email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			serv.log.Warn("login failed: user not found", zap.String("email", req.Email))
			return nil, &apperror.APIError{
				Code:    "USER_NOT_FOUND",
				Message: "No account found with this email.",
			}
		}

		serv.log.Error("failed to fetch user from db",
			zap.String("email", req.Email),
			zap.Error(err),
		)
		return nil, &apperror.APIError{
			Code:    "DB_ERROR",
			Message: "Something went wrong. Please try again later.",
		}
	}

	if fetchedUser.Status == "blocked" || fetchedUser.Status == "deleted" {
		serv.log.Warn("blocked or deleted user attempted login", zap.Int64("user_id", fetchedUser.ID))
		return nil, &apperror.APIError{
			Code:    "USER_BLOCKED",
			Message: "Your account is not active. Please contact support.",
		}
	}

	if !utils.VerifyPassword(fetchedUser.Password, req.Password) {
		serv.log.Warn("wrong password", zap.String("email", req.Email))
		return nil, &apperror.APIError{
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
			Code:    "JWT_GENERATION_FAILED",
			Message: "Could not generate token. Please try again.",
		}
	}
	var expiryAt time.Time
	resp.RefreshToken, expiryAt, err = utils.GenerateTokenWithExpiry(32, 15*1440)
	if err != nil {
		serv.log.Error("Failed to generate refresh token", zap.String("service", "UserService"), zap.Error(err))
		return nil, &apperror.APIError{
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
			Code:    "USER_NOT_FOUND",
			Message: "No account found with this email.",
		}
	}
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, &apperror.APIError{
				Code:    "USER_NOT_FOUND",
				Message: err.Error(),
			}
		}
	}

	authTokens := domain.AuthTokens{}
	authTokens.AccessToken, err = utils.IssueJWT(user.ID, user.Email, user.Role, user.IsVerified, serv.log)
	if err != nil {
		serv.log.Error("Failed to issue jwt", zap.String("service", "UserService"), zap.Error(err))
		return nil, &apperror.APIError{
			Code:    "JWT_GENERATION_FAILED",
			Message: "Could not generate token. Please try again.",
		}
	}
	authTokens.RefreshToken, err = utils.GenerateToken(32)
	if err != nil {
		serv.log.Error("Failed to generate refresh token", zap.String("service", "UserService"), zap.Error(err))
		return nil, &apperror.APIError{
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
func (serv *UserService) GetUserProfile(ctx context.Context, userID int64) (*domain.UserProfile, *apperror.APIError) {

	user, err := serv.repo.GetUserByID(ctx, userID)
	if err != nil {
		serv.log.Debug("failed to fetch user profile from db", zap.Int64("user_id", userID), zap.Error(err))
		return nil, &apperror.APIError{
			Code:    "DB_ERROR",
			Message: "Failed to fetch user profile.",
		}
	}

	addresses, err := serv.repo.GetUserAddresses(ctx, userID)
	if err != nil {
		serv.log.Debug("failed to fetch user addresses from db", zap.Int64("user_id", userID), zap.Error(err))
		return nil, &apperror.APIError{
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
		DefaultAddressID: user.DefaultAddressID,
		Status:           &user.Status,
		CreatedAt:        &user.CreatedAt,
		Addresses:        addresses,
	}

	return &userProfile, nil
}

// update user profile
func (serv *UserService) UpdateUserProfile(ctx context.Context, req *domain.UpdateUserProfileRequest) (*domain.UserProfile, *apperror.APIError) {

	userID, err := utils.GetUserIDFromContext(ctx)
	if err != nil {
		return nil, &apperror.APIError{
			Code:    "INVALID_USER_ID",
			Message: "Invalid user ID.",
		}
	}
	updatedProfile, err := serv.repo.UpdateUserProfile(ctx, userID, req)
	if err != nil {
		if err == apperror.ErrPhoneExists {
			return nil, &apperror.APIError{
				Code:    "PHONE_ALREADY_EXISTS",
				Message: "A user with this phone number already exists.",
			}
		}
		serv.log.Debug("failed to update user profile in db", zap.Int64("user_id", userID), zap.Error(err))
		return nil, &apperror.APIError{
			Code:    "DB_ERROR",
			Message: "Failed to update user profile.",
		}
	}
	return updatedProfile, nil
}

// add user address
func (serv *UserService) AddUserAddress(ctx context.Context, address *domain.UserAddress) *apperror.APIError {
	err := serv.repo.InsertUserAddress(ctx, address)
	if err != nil {
		serv.log.Debug("failed to insert user address into db", zap.Int64("user_id", address.UserID), zap.Error(err))
		return &apperror.APIError{
			Code:    "DB_ERROR",
			Message: "Failed to add user address.",
		}
	}
	return nil
}

// get user addresses
func (serv *UserService) GetUserAddresses(ctx context.Context, userID int64) ([]*domain.UserAddress, *apperror.APIError) {
	addresses, err := serv.repo.GetUserAddresses(ctx, userID)
	if err != nil {
		serv.log.Debug("failed to fetch user addresses from db", zap.Int64("user_id", userID), zap.Error(err))
		return nil, &apperror.APIError{
			Code:    "DB_ERROR",
			Message: "Failed to fetch user addresses.",
		}
	}
	return addresses, nil
}
