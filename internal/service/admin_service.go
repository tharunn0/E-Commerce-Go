package service

import (
	"context"
	"errors"
	"os"

	"github.com/jackc/pgx/v5"
	"github.com/tharunn0/E-Commerce-Go/internal/apperror"
	"github.com/tharunn0/E-Commerce-Go/internal/domain"
	"github.com/tharunn0/E-Commerce-Go/internal/utils"
	"go.uber.org/zap"
)

type AdminService struct {
	repo domain.AdminRepository
	log  *zap.Logger
}

func NewAdminService(adminRepo domain.AdminRepository, logger *zap.Logger) *AdminService {
	return &AdminService{
		repo: adminRepo,
		log:  logger,
	}
}

func (serv *AdminService) RegisterAdmin(ctx context.Context, req *domain.AdminRegisterRequest) *apperror.APIError {

	if !utils.IsValidEmail(req.Email) {
		return &apperror.APIError{
			Code:    "INVALID_EMAIL",
			Message: "Please provide a valid email address.",
		}
	}

	if !utils.IsValidPassword(req.Password) {
		return &apperror.APIError{
			Code:    "INVALID_PASSWORD",
			Message: "Please provide a valid password.",
		}
	}

	hashedPass := utils.HashPassword(req.Password)
	if len(hashedPass) == 0 {
		return &apperror.APIError{
			Code:    "HASHING_FAILED",
			Message: "Could not process password. Please try again.",
		}
	}
	req.Password = hashedPass

	if req.AdminCode != os.Getenv("ADMIN_CODE") {
		return &apperror.APIError{
			Code:    "INVALID_ADMIN_CODE",
			Message: "Invalid admin code. Please try again.",
		}
	}

	err := serv.repo.CreateAdmin(ctx, req)
	if err != nil {
		if errors.Is(err, apperror.ErrEmailExists) {
			return &apperror.APIError{
				Code:    "EMAIL_ALREADY_EXISTS",
				Message: "An account with this email already exists.",
			}
		}
		if errors.Is(err, apperror.ErrPhoneExists) {
			return &apperror.APIError{
				Code:    "PHONE_ALREADY_EXISTS",
				Message: "An account with this phone number already exists.",
			}
		}
		serv.log.Error("failed to create admin", zap.String("email", req.Email), zap.Error(err))
		return &apperror.APIError{
			Code:    "DB_ERROR",
			Message: "Something went wrong. Please try again later.",
		}
	}

	return nil

}

func (serv *AdminService) LoginAdmin(req *domain.LoginRequest) (*domain.LoginResponse, *apperror.APIError) {
	ctx := context.Background()

	if !utils.IsValidEmail(req.Email) {
		return nil, &apperror.APIError{
			Code:    "INVALID_EMAIL",
			Message: "Please provide a valid email address.",
		}
	}

	fetchedUser, err := serv.repo.GetAdmin(ctx, req.Email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, &apperror.APIError{
				Code:    "ADMIN_NOT_FOUND",
				Message: "No account found with this email.",
			}
		}

		return nil, &apperror.APIError{
			Code:    "DB_ERROR",
			Message: "Something went wrong. Please try again later.",
		}
	}

	if fetchedUser.Role != "admin" {
		return nil, &apperror.APIError{
			Code:    "ACCESS_DENIED",
			Message: "Only admin accounts can log in here.",
		}
	}

	if !utils.VerifyPassword(fetchedUser.Password, req.Password) {
		return nil, &apperror.APIError{
			Code:    "WRONG_PASSWORD",
			Message: "Invalid email or password.",
		}
	}
	resp := &domain.LoginResponse{}

	resp.User.ID = fetchedUser.ID
	resp.User.Email = fetchedUser.Email
	resp.User.FirstName = fetchedUser.FirstName
	resp.User.LastName = fetchedUser.LastName
	resp.User.Role = string(fetchedUser.Role)

	// Issue JWT
	resp.AccessToken, err = utils.IssueJWT(fetchedUser.ID, fetchedUser.Email, fetchedUser.Role, fetchedUser.IsVerified, serv.log)
	if resp.AccessToken == "" || err != nil {
		serv.log.Error("failed to issue jwt", zap.Error(err))
		return nil, &apperror.APIError{
			Code:    "TOKEN_GENERATION_FAILED",
			Message: "Could not generate access token.",
		}
	}

	resp.RefreshToken, err = utils.GenerateToken(32)
	if resp.RefreshToken == "" || err != nil {
		serv.log.Error("failed to generate refresh token", zap.Error(err))
		return nil, &apperror.APIError{
			Code:    "TOKEN_GENERATION_FAILED",
			Message: "Could not generate refresh token.",
		}
	}

	return resp, nil
}

func (serv *AdminService) GetAllUsers(ctx context.Context, req *domain.UserFilter) ([]*domain.UserProfile, *apperror.APIError) {
	users, err := serv.repo.ListUsers(ctx, req)
	if err != nil {
		serv.log.Error("failed to retrive users", zap.String("Function", "repo.ListUsers"), zap.Error(err))
		return nil, &apperror.APIError{
			Code:    "",
			Message: "Failed to retrieve users",
		}
	}

	return users, nil
}

func (serv *AdminService) UpdateUserStatus(ctx context.Context, req *domain.UserStatusUpdateRequest) *apperror.APIError {
	if !utils.IsValidStatus(req.Status) {
		return &apperror.APIError{
			Code:    "INVALID_STATUS",
			Message: "Invalid status. Please provide a valid status.",
		}
	}

	if req.UserID <= 0 {
		return &apperror.APIError{
			Code:    "INVALID_USER_ID",
			Message: "Invalid user ID. Please provide a valid user ID.",
		}
	}

	err := serv.repo.UpdateUserStatus(ctx, req)
	if err != nil {
		return &apperror.APIError{
			Code:    "DB_ERROR",
			Message: "Something went wrong. Please try again later.",
		}
	}
	return nil
}

func (serv *AdminService) GetUserByID(ctx context.Context, id int64) (*domain.UserProfile, *apperror.APIError) {

	if id <= 0 {
		return nil, &apperror.APIError{
			Code:    "INVALID_USER_ID",
			Message: "Invalid user ID. Please provide a valid user ID.",
		}
	}

	user, err := serv.repo.GetUsersByID(ctx, id)
	if err != nil {
		serv.log.Warn("")
		return nil, &apperror.APIError{
			Code:    "DB_ERROR",
			Message: "Failed to fetch user",
		}
	}

	return user, nil

}

func (serv *AdminService) DeleteUser(ctx context.Context, id int64) *apperror.APIError {

	if id <= 0 {
		return &apperror.APIError{
			Code:    "INVALID_USER_ID",
			Message: "Invalid user ID. Please provide a valid user ID.",
		}
	}

	err := serv.repo.DeleteUser(ctx, id)
	if err != nil {
		serv.log.Warn("")
		return &apperror.APIError{
			Code:    "DB_ERROR",
			Message: "Failed to delte user",
		}
	}

	return nil

}
