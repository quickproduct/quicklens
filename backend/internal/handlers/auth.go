package handlers

import (
	"go.uber.org/zap"
	"net/http"
	"time"

	"github.com/google/uuid"

	"quicklens/backend/internal/auth"
	"quicklens/backend/internal/db"
	"quicklens/backend/internal/models"
)

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	var req models.LoginRequest
	if err := ParseJSON(r, &req); err != nil {
		zap.L().Sugar().Infof("Login: Failed to parse JSON request body: %v", err)
		WriteError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Email == "" || req.Password == "" {
		zap.L().Sugar().Infof("Login: Missing email or password in request")
		WriteError(w, http.StatusBadRequest, "Email and password are required")
		return
	}

	zap.L().Sugar().Infof("Login attempt: email=%q", req.Email)

	var user models.User
	err := db.DB.QueryRow(
		"SELECT id, email, hashed_password, role, is_active FROM users WHERE email = ?",
		req.Email,
	).Scan(&user.ID, &user.Email, &user.HashedPassword, &user.Role, &user.IsActive)
	if err != nil {
		zap.L().Sugar().Infof("Login failed: email %q not found in DB or error: %v", req.Email, err)
		WriteError(w, http.StatusUnauthorized, "Invalid email or password")
		return
	}

	if !user.IsActive {
		zap.L().Sugar().Infof("Login failed: account %q is disabled", req.Email)
		WriteError(w, http.StatusForbidden, "Account is disabled")
		return
	}

	if !auth.CheckPasswordHash(req.Password, user.HashedPassword) {
		zap.L().Sugar().Infof("Login failed: password check failed for %q (hashed_pwd: %s)", req.Email, user.HashedPassword)
		WriteError(w, http.StatusUnauthorized, "Invalid email or password")
		return
	}

	zap.L().Sugar().Infof("Login success: email=%q role=%q ID=%s", req.Email, user.Role, user.ID)

	accessToken, err := auth.GenerateAccessToken(user.ID)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "Failed to generate access token")
		return
	}

	refreshToken, err := auth.GenerateRefreshToken(user.ID)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "Failed to generate refresh token")
		return
	}

	// Store session
	sessionID := uuid.New().String()
	expiresAt := time.Now().Add(7 * 24 * time.Hour)
	_, err = db.DB.Exec(
		"INSERT INTO sessions (id, user_id, refresh_token, expires_at, created_at) VALUES (?, ?, ?, ?, ?)",
		sessionID, user.ID, refreshToken, expiresAt, time.Now().UTC(),
	)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "Failed to create session")
		return
	}

	WriteJSON(w, http.StatusOK, models.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
	})
}

func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	var req models.RefreshRequest
	if err := ParseJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	_, err := db.DB.Exec("DELETE FROM sessions WHERE refresh_token = ?", req.RefreshToken)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "Failed to logout")
		return
	}

	WriteJSON(w, http.StatusOK, map[string]string{"message": "Logged out successfully"})
}

func RefreshHandler(w http.ResponseWriter, r *http.Request) {
	var req models.RefreshRequest
	if err := ParseJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.RefreshToken == "" {
		WriteError(w, http.StatusBadRequest, "Refresh token is required")
		return
	}

	claims, err := auth.VerifyToken(req.RefreshToken, "refresh")
	if err != nil {
		WriteError(w, http.StatusUnauthorized, "Invalid or expired refresh token")
		return
	}

	// Verify session exists in DB
	var sessionID string
	err = db.DB.QueryRow(
		"SELECT id FROM sessions WHERE refresh_token = ? AND user_id = ? AND expires_at > ?",
		req.RefreshToken, claims.Sub, time.Now().UTC(),
	).Scan(&sessionID)
	if err != nil {
		WriteError(w, http.StatusUnauthorized, "Session not found or expired")
		return
	}

	// Generate new tokens
	accessToken, err := auth.GenerateAccessToken(claims.Sub)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "Failed to generate access token")
		return
	}

	refreshToken, err := auth.GenerateRefreshToken(claims.Sub)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "Failed to generate refresh token")
		return
	}

	// Update session with new refresh token
	newExpiry := time.Now().Add(7 * 24 * time.Hour)
	_, _ = db.DB.Exec(
		"UPDATE sessions SET refresh_token = ?, expires_at = ? WHERE id = ?",
		refreshToken, newExpiry, sessionID,
	)

	WriteJSON(w, http.StatusOK, models.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
	})
}

func MeHandler(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(UserIDKey).(string)

	var user models.UserResponse
	var createdAt time.Time
	err := db.DB.QueryRow(
		"SELECT id, email, role, is_active, created_at FROM users WHERE id = ?",
		userID,
	).Scan(&user.ID, &user.Email, &user.Role, &user.IsActive, &createdAt)
	if err != nil {
		WriteError(w, http.StatusNotFound, "User not found")
		return
	}
	user.CreatedAt = &createdAt

	WriteJSON(w, http.StatusOK, user)
}

func ChangePasswordHandler(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(UserIDKey).(string)

	var req models.ChangePasswordRequest
	if err := ParseJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.CurrentPassword == "" || req.NewPassword == "" {
		WriteError(w, http.StatusBadRequest, "Current and new passwords are required")
		return
	}

	if len(req.NewPassword) < 6 {
		WriteError(w, http.StatusBadRequest, "New password must be at least 6 characters")
		return
	}

	var hashedPassword string
	err := db.DB.QueryRow("SELECT hashed_password FROM users WHERE id = ?", userID).Scan(&hashedPassword)
	if err != nil {
		WriteError(w, http.StatusNotFound, "User not found")
		return
	}

	if !auth.CheckPasswordHash(req.CurrentPassword, hashedPassword) {
		WriteError(w, http.StatusUnauthorized, "Current password is incorrect")
		return
	}

	newHash, err := auth.HashPassword(req.NewPassword)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "Failed to hash password")
		return
	}

	_, err = db.DB.Exec(
		"UPDATE users SET hashed_password = ?, updated_at = ? WHERE id = ?",
		newHash, time.Now().UTC(), userID,
	)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "Failed to update password")
		return
	}

	WriteJSON(w, http.StatusOK, map[string]string{"message": "Password updated successfully"})
}

func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	var req models.RegisterRequest
	if err := ParseJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Email == "" || req.Password == "" {
		WriteError(w, http.StatusBadRequest, "Email and password are required")
		return
	}

	if len(req.Password) < 6 {
		WriteError(w, http.StatusBadRequest, "Password must be at least 6 characters")
		return
	}

	// Check if email already exists
	var exists int
	err := db.DB.QueryRow("SELECT COUNT(*) FROM users WHERE email = ?", req.Email).Scan(&exists)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "Database error")
		return
	}
	if exists > 0 {
		WriteError(w, http.StatusConflict, "Email already registered")
		return
	}

	hashedPassword, err := auth.HashPassword(req.Password)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "Failed to hash password")
		return
	}

	id := uuid.New().String()
	now := time.Now().UTC()

	_, err = db.DB.Exec(
		"INSERT INTO users (id, email, hashed_password, role, is_active, created_at, updated_at) VALUES (?, ?, ?, 'user', 1, ?, ?)",
		id, req.Email, hashedPassword, now, now,
	)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "Failed to create user")
		return
	}

	WriteJSON(w, http.StatusCreated, models.UserResponse{
		ID:        id,
		Email:     req.Email,
		Role:      "user",
		IsActive:  true,
		CreatedAt: &now,
	})
}
