package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
	"personal-assistant-backend/internal/models"
)

// setupLoginRouter initializes router + AuthHandler with mock dependencies
func setupLoginRouter(t *testing.T) (*gin.Engine, sqlmock.Sqlmock) {
	gin.SetMode(gin.TestMode)

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock: %v", err)
	}

	h := &AuthHandler{
		db: db,
		generateJWT: func(userID string, _ time.Duration) (string, error) {
			if userID == "fail-token" {
				return "", errors.New("token error")
			}
			return "mock-token-" + userID, nil
		},
		getAccessTTL:  func() time.Duration { return 15 * time.Minute },
		getRefreshTTL: func() time.Duration { return 30 * 24 * time.Hour },
	}

	r := gin.Default()
	r.POST("/login", h.Login)

	return r, mock
}

// --- tests ---

func TestLogin_Success(t *testing.T) {
	router, mock := setupLoginRouter(t)

	password := "supersecret"
	hash, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	now := time.Now()

	mock.ExpectQuery(`SELECT id, first_name, last_name, email, phone_number, password_hash, created_at FROM users WHERE email=\$1`).
		WithArgs("jane@example.com").
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "first_name", "last_name", "email", "phone_number", "password_hash", "created_at",
		}).AddRow("123", "Jane", "Doe", "jane@example.com", "555-1234", string(hash), now))

	body := `{"email":"jane@example.com","password":"supersecret"}`
	req, _ := http.NewRequest("POST", "/login", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp models.AuthWithTokensResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, "Jane", resp.User.FirstName)
	assert.Equal(t, "mock-token-123", resp.AccessToken)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestLogin_InvalidPassword(t *testing.T) {
	router, mock := setupLoginRouter(t)

	hash, _ := bcrypt.GenerateFromPassword([]byte("differentpass"), bcrypt.DefaultCost)
	now := time.Now()

	mock.ExpectQuery(`SELECT id, first_name, last_name, email, phone_number, password_hash, created_at FROM users WHERE email=\$1`).
		WithArgs("bob@example.com").
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "first_name", "last_name", "email", "phone_number", "password_hash", "created_at",
		}).AddRow("456", "Bob", "Smith", "bob@example.com", "555-4567", string(hash), now))

	body := `{"email":"bob@example.com","password":"wrongpass"}`
	req, _ := http.NewRequest("POST", "/login", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "invalid credentials")
}

func TestLogin_UserNotFound(t *testing.T) {
	router, mock := setupLoginRouter(t)

	mock.ExpectQuery(`SELECT id, first_name, last_name, email, phone_number, password_hash, created_at FROM users WHERE email=\$1`).
		WithArgs("ghost@example.com").
		WillReturnError(sql.ErrNoRows)

	body := `{"email":"ghost@example.com","password":"irrelevant"}`
	req, _ := http.NewRequest("POST", "/login", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "invalid credentials")
}

func TestLogin_DBError(t *testing.T) {
	router, mock := setupLoginRouter(t)

	mock.ExpectQuery(`SELECT id, first_name, last_name, email, phone_number, password_hash, created_at FROM users WHERE email=\$1`).
		WithArgs("crash@example.com").
		WillReturnError(errors.New("db exploded"))

	body := `{"email":"crash@example.com","password":"anypass"}`
	req, _ := http.NewRequest("POST", "/login", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "db error")
}

func TestLogin_BadRequest(t *testing.T) {
	router, _ := setupLoginRouter(t)

	req, _ := http.NewRequest("POST", "/login", strings.NewReader(`{invalid-json}`))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "invalid payload")
}
