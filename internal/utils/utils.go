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
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/tharunn0/E-Commerce-Go/internal/domain"
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

func HashOTP(otp string) string {
	res, er := bcrypt.GenerateFromPassword([]byte(otp), bcrypt.DefaultCost)
	if er != nil {
		return ""
	}
	return string(res)
}
func VerifyOTP(hash, otp string) bool {
	er := bcrypt.CompareHashAndPassword([]byte(hash), []byte(otp))
	return er == nil
}

func GenerateState() string {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return ""
	}
	return base64.URLEncoding.EncodeToString(b)
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

	secretkey := []byte(os.Getenv("JWT_SECRET"))
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

func GetEmailFromContext(ctx context.Context) string {
	if email, ok := ctx.Value(domain.KeyEmail).(string); ok {
		return email
	}
	return ""
}

func GetUserIDFromContext(ctx context.Context) (int64, error) {
	if userID, ok := ctx.Value(domain.KeyUserID).(float64); ok {
		return int64(userID), nil
	}
	return 0, fmt.Errorf("user ID not found in context")
}

func ValidateRefreshToken(token *domain.RefreshToken) error {

	fmt.Println("token", token)
	if time.Now().After(token.ExpiryAt) {
		return fmt.Errorf("refresh token expired")
	}
	if token.Revoked {
		return fmt.Errorf("refresh token revoked")
	}
	return nil
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

func GenerateTokenWithExpiry(len int, timeInMinutes int) (string, time.Time, error) {
	if len < 8 {
		len = 8
	}
	if timeInMinutes < 1 {
		timeInMinutes = 1
	}
	token, err := GenerateToken(len)
	if err != nil {
		return "", time.Time{}, err
	}
	expiryAt := time.Now().Add(time.Duration(timeInMinutes) * time.Minute)
	return token, expiryAt, nil
}
func GetUserRole(ctx context.Context) string {
	if role, ok := ctx.Value(domain.KeyRole).(string); ok {
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

func IsValidStatus(status string) bool {
	return status == "active" || status == "blocked" || status == "deleted"
}

func IsFiltersValid(f *domain.ProductFilter) bool {

	if f.Sort == "price" {
		f.Sort = "base_price"
	}

	if f.Sort == "" || f.Order == "" {
		return true
	}

	f.Order = strings.ToUpper(f.Order)
	f.Sort = strings.ToLower(f.Sort)

	validOrder := map[string]struct{}{
		"ASC":  {},
		"DESC": {},
	}
	validSortCol := map[string]struct{}{
		"base_price": {},
		"name":       {},
		"created_at": {},
		"rating":     {},
	}

	if _, ok := validOrder[f.Order]; !ok {
		return false
	}
	if _, ok := validSortCol[f.Sort]; !ok {
		return false
	}

	return true
}

func ExtractAuthToken(rawtoken string) (string, error) {

	if rawtoken == "" {
		return "", fmt.Errorf("authorization header missing")
	}

	parts := strings.SplitN(rawtoken, " ", 2)
	if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
		return "", fmt.Errorf("invalid authorization header format")
	}
	return parts[1], nil
}

func ValidateUserAddress(address *domain.UserAddress) error {
	if address.Label == "" {
		return fmt.Errorf("Address label is required")
	}
	if address.AddressLine == "" {
		return fmt.Errorf("Address line is required")
	}
	if address.District == "" {
		return fmt.Errorf("District is required")
	}
	if address.Pincode == "" {
		return fmt.Errorf("Pincode is required")
	}
	if address.City == "" {
		return fmt.Errorf("City is required")
	}
	if address.State == "" {
		return fmt.Errorf("State is required")
	}
	if address.Country == "" {
		return fmt.Errorf("Country is required")
	}
	return nil
}
