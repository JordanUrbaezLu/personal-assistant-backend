package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgconn"
	"github.com/stretchr/testify/assert"
	"personal-assistant-backend/internal/models"
)

// --- mock helpers ---
var mockSignupGenerateJWT = func(userID string, _ time.Duration) (string, error) {
	if userID == "fail-token" {
		return "", errors.New("token generation failed")
	}
	return "mock-token-" + userID, nil
}

var mockSignupGetAccessTTL = func() time.Duration { return 15 * time.Minute }
var mockSignupGetRefreshTTL = func() time.Duration { return 30 * 24 * time.Hour }

// --- setup ---
func setupSignupRouter(t *testing.T) (*gin.Engine, sqlmock.Sqlmock) {
	gin.SetMode(gin.TestMode)

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock: %v", err)
	}

	h := &AuthHandler{
		db:            db,
		generateJWT:   mockSignupGenerateJWT,
		getAccessTTL:  mockSignupGetAccessTTL,
		getRefreshTTL: mockSignupGetRefreshTTL,
	}

	r := gin.Default()
	r.POST("/signup", h.Signup)
	return r, mock
}

// --- tests ---

func TestSignup_Success(t *testing.T) {
	router, mock := setupSignupRouter(t)

	now := time.Now()

	mock.ExpectQuery(`INSERT INTO users \(first_name, last_name, email, password_hash, phone_number, created_at\) VALUES \(\$1, \$2, \$3, \$4, \$5, \$6\) RETURNING id, created_at`).
		WithArgs("Jane", "Doe", "jane@example.com", sqlmock.AnyArg(), "555-1234", sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"id", "created_at"}).AddRow("123", now))

	body := `{
		"first_name": "Jane",
		"last_name": "Doe",
		"email": "jane@example.com",
		"password": "supersecret",
		"phone_number": "555-1234"
	}`

	req, _ := http.NewRequest("POST", "/signup", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var resp models.AuthWithTokensResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, "Jane", resp.User.FirstName)
	assert.Equal(t, "mock-token-123", resp.AccessToken)
	assert.Equal(t, "mock-token-123", resp.RefreshToken)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestSignup_InvalidPayload(t *testing.T) {
	router, _ := setupSignupRouter(t)

	req, _ := http.NewRequest("POST", "/signup", strings.NewReader(`{invalid-json}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "invalid payload")
}

func TestSignup_DBConflict(t *testing.T) {
	router, mock := setupSignupRouter(t)

	mock.ExpectQuery(`INSERT INTO users`).WillReturnError(&pgconn.PgError{Code: "23505"})

	body := `{
		"first_name": "Bob",
		"last_name": "Smith",
		"email": "bob@example.com",
		"password": "password123",
		"phone_number": "555-9876"
	}`

	req, _ := http.NewRequest("POST", "/signup", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusConflict, w.Code)
	assert.Contains(t, w.Body.String(), "email already exists")
}

func TestSignup_DBError(t *testing.T) {
	router, mock := setupSignupRouter(t)

	mock.ExpectQuery(`INSERT INTO users`).WillReturnError(errors.New("db exploded"))

	body := `{
		"first_name": "Alice",
		"last_name": "Brown",
		"email": "alice@example.com",
		"password": "testpass",
		"phone_number": "555-0000"
	}`

	req, _ := http.NewRequest("POST", "/signup", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "db error")
}

func TestSignup_HashError(t *testing.T) {
	// ✅ Override bcryptGenerate before router setup
	oldBcrypt := bcryptGenerate
	bcryptGenerate = func(_ []byte, _ int) ([]byte, error) {
		return nil, errors.New("hash fail")
	}
	defer func() { bcryptGenerate = oldBcrypt }()

	router, _ := setupSignupRouter(t)

	body := `{
		"first_name": "Fail",
		"last_name": "Hash",
		"email": "fail@example.com",
		"password": "longenough",
		"phone_number": "000-0000"
	}`

	req, _ := http.NewRequest("POST", "/signup", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "failed to hash password")
}


