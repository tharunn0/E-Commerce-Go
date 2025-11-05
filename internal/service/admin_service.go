package service

import (
	"context"
	"errors"

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

	// Issue JWT
	token, err := utils.IssueJWT(fetchedUser.ID, fetchedUser.Email, fetchedUser.Role, fetchedUser.IsVerified, serv.log)
	if len(token) == 0 || err != nil {
		serv.log.Error("failed to issue jwt", zap.Error(err))
		return nil, &apperror.APIError{
			Code:    "TOKEN_GENERATION_FAILED",
			Message: "Could not generate access token.",
		}
	}

	resp := &domain.LoginResponse{
		Token: token,
	}
	resp.User.ID = fetchedUser.ID
	resp.User.Email = fetchedUser.Email
	resp.User.FirstName = fetchedUser.FirstName
	resp.User.LastName = fetchedUser.LastName
	resp.User.Role = string(fetchedUser.Role)

	return resp, nil
}

func (serv *AdminService) GetAllUsers(ctx context.Context, req *domain.UserFilter) ([]*domain.User, *apperror.APIError) {
	users, err := serv.repo.ListUsers(ctx, req)
	if err != nil {
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
