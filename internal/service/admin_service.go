package service

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/tharunn0/E-Commerce-Go/internal/domain"
	"github.com/tharunn0/E-Commerce-Go/internal/infrastructure/repository"
	"github.com/tharunn0/E-Commerce-Go/internal/utils"
	"go.uber.org/zap"
)

type AdminService struct {
	repo repository.AdminRepository
	log  *zap.Logger
}

func NewAdminService(adminRepo repository.AdminRepository, logger *zap.Logger) *AdminService {
	return &AdminService{
		repo: adminRepo,
		log:  logger,
	}
}

func (serv *AdminService) LoginAdmin(req *domain.LoginRequest) (*domain.LoginResponse, *domain.APIError) {

	ctx := context.Background()
	serv.log.Info("login attempt started", zap.String("email", req.Email))

	if !utils.IsValidEmail(req.Email) {
		serv.log.Warn("invalid email format", zap.String("email", req.Email))
		return nil, &domain.APIError{
			Code:    "INVALID_EMAIL",
			Message: "Please provide a valid email address.",
		}
	}

	fetchedUser, err := serv.repo.GetAdmin(ctx, req.Email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			serv.log.Warn("login failed: user not found", zap.String("email", req.Email))
			return nil, &domain.APIError{
				Code:    "ADMIN_NOT_FOUND",
				Message: "No account found with this email.",
			}
		}

		serv.log.Error("failed to fetch user from db",
			zap.String("email", req.Email),
			zap.Error(err),
		)
		return nil, &domain.APIError{
			Code:    "DB_ERROR",
			Message: "Something went wrong. Please try again later.",
		}
	}

	if fetchedUser.Role != "admin" {
		serv.log.Warn("non-admin user attempted to access admin login",
			zap.Int64("user_id", fetchedUser.ID),
			zap.String("user_email", fetchedUser.Email),
			zap.String("user_role", fetchedUser.Role),
		)

		return nil, &domain.APIError{
			Code:    "ACCESS_DENIED",
			Message: "Only admin accounts can log in here.",
		}
	}

	if !utils.VerifyPassword(fetchedUser.Password, req.Password) {
		serv.log.Warn("wrong password", zap.String("email", req.Email))
		return nil, &domain.APIError{
			Code:    "WRONG_PASSWORD",
			Message: "Invalid email or password.",
		}
	}

	// Issue JWT
	token, err := utils.IssueJWT(fetchedUser.ID, fetchedUser.Email, fetchedUser.Role, fetchedUser.IsVerified, serv.log)
	if len(token) == 0 {
		serv.log.Error("failed to issue jwt", zap.Error(err))
		return nil, &domain.APIError{
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

	serv.log.Info("login successful",
		zap.Int64("user_id", fetchedUser.ID),
		zap.String("email", fetchedUser.Email),
		zap.String("role", string(fetchedUser.Role)),
	)

	return resp, nil
}
