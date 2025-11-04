package domain

type contextKey string

const (
	KeyUserID   contextKey = "userId"
	KeyRole     contextKey = "role"
	KeyVerified contextKey = "verified"
	KeyEmail    contextKey = "email"
)
