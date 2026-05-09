package gateway

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type AuthHandler struct {
	db *sql.DB
}

func NewAuthHandler(db *sql.DB) *AuthHandler {
	return &AuthHandler{db: db}
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type loginResponse struct {
	Token string `json:"token"`
	User  any    `json:"user"`
}

type userRow struct {
	ID       string
	Name     string
	Email    string
	Password string
	Role     string
}

// ─────────────────────────────────────────────────────────────
// POST /auth/login
// ─────────────────────────────────────────────────────────────

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req loginRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid body")
		return
	}

	user, err := h.findUserByEmail(ctx, req.Email)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "invalid credentials")
		return
	}

	if err := bcrypt.CompareHashAndPassword(
		[]byte(user.Password),
		[]byte(req.Password),
	); err != nil {
		writeError(w, http.StatusUnauthorized, "invalid credentials")
		return
	}

	token, err := generateJWT(user)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed generating token")
		return
	}

	writeJSON(w, http.StatusOK, loginResponse{
		Token: token,
		User: map[string]any{
			"id":    user.ID,
			"name":  user.Name,
			"email": user.Email,
			"role":  user.Role,
		},
	})
}

func (h *AuthHandler) findUserByEmail(ctx context.Context, email string) (userRow, error) {
	var u userRow

	err := h.db.QueryRowContext(ctx, `
		SELECT id, name, email, password, role
		FROM users
		WHERE email = ?
		LIMIT 1
	`, email).Scan(
		&u.ID,
		&u.Name,
		&u.Email,
		&u.Password,
		&u.Role,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return userRow{}, errors.New("user not found")
		}
		return userRow{}, err
	}

	return u, nil
}

// ─────────────────────────────────────────────────────────────
// JWT
// ─────────────────────────────────────────────────────────────

func generateJWT(user userRow) (string, error) {
	secret := os.Getenv("JWT_SECRET")

	claims := jwt.MapClaims{
		"sub":   user.ID,
		"name":  user.Name,
		"email": user.Email,
		"role":  user.Role,
		"exp":   time.Now().Add(24 * time.Hour).Unix(),
		"iat":   time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString([]byte(secret))
}

// ─────────────────────────────────────────────────────────────
// HELPERS
// ─────────────────────────────────────────────────────────────

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]any{
		"error": message,
	})
}