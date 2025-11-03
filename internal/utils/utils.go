package utils

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"math/big"
	"os"
	"regexp"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

func HashPassword(pass string) string {
	res, er := bcrypt.GenerateFromPassword([]byte(pass), bcrypt.DefaultCost)
	if er != nil {
		return ""
	}
	return string(res)
}

func VerifyPassword(hash, pass string) bool {
	er := bcrypt.CompareHashAndPassword([]byte(hash), []byte(pass))
	return er == nil
}

func IsValidEmail(email string) bool {
	return emailRegex.MatchString(email)
}

func IsValidPassword(password string) bool {
	if len(password) < 8 {
		return false
	}

	hasLower := regexp.MustCompile(`[a-z]`).MatchString(password)
	hasUpper := regexp.MustCompile(`[A-Z]`).MatchString(password)
	hasDigit := regexp.MustCompile(`\d`).MatchString(password)
	hasSpecial := regexp.MustCompile(`[\W_]`).MatchString(password)

	return hasLower && hasUpper && hasDigit && hasSpecial
}

func IssueJWT(uid int64, email, role string, isverified bool, log *zap.Logger) (string, error) {

	secretkey := []byte(os.Getenv("HS_256KEY"))
	if len(secretkey) == 0 {
		log.Error("SECRET_KEY_NOT_FOUND_IN_ENVIRONMENT", zap.String("service", "utils"))
		return "", fmt.Errorf("internal server configuration error")
	}

	expiryTime, err := strconv.Atoi(os.Getenv("JWT_EXPIRY_TIME"))
	if err != nil {
		log.Error("JWT_EXPIRY_TIME_NOT_FOUND_IN_ENVIRONMENT", zap.String("service", "utils"))
		return "", fmt.Errorf("internal server configuration error")
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id":  uid,
		"email":    email,
		"role":     role,
		"verified": isverified,
		"iat":      time.Now().Unix(),
		"exp":      time.Now().Add(time.Duration(expiryTime) * time.Minute).Unix(),
	})
	str, er := token.SignedString(secretkey)
	if er != nil || str == "" {
		log.Error("FAILED_TO_ISSUE_JWT", zap.String("service", "utils"))
		return "", er
	}
	return str, nil

}

func GenerateOTP() (string, error) {
	max := big.NewInt(900000)

	n, err := rand.Int(rand.Reader, max)
	if err != nil {
		return "", err
	}
	otp := n.Int64() + 100000
	return fmt.Sprintf("%06d", otp), nil

}

func GenerateToken(len int) (string, error) {

	if len < 8 {
		len = 8
	}

	b := make([]byte, len)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(b), nil
}

func GetUserRole(ctx context.Context) string {
	if role, ok := ctx.Value("role").(string); ok {
		return role
	}
	return "user"
}

func IsAdmin(ctx context.Context) bool {
	return GetUserRole(ctx) == "admin"
}

// func ValidateGoogleSignIntoken(ctx context.Context, token string) {
// 	idtoken.Validate(ctx, token, "")
// }

func IsValidDataType(dataType string) bool {
	return dataType == "text" || dataType == "number" || dataType == "boolean" || dataType == "date"
}
