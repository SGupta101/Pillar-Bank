package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
	"github.com/stretchr/testify/assert"
)

func TestCreateToken(t *testing.T) {
	token, err := CreateToken("testuser")
	assert.NoError(t, err)
	assert.NotEmpty(t, token)
}

// tests AuthenticateMiddleware with valid and invalid tokens
func TestAuthenticateMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.Default()

	router.GET("/test", AuthenticateMiddleware, func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "Authenticated"})
	})

	tests := []struct {
		name         string
		setupCookie  func() *http.Cookie
		expectedCode int
		expectedBody string
	}{
		{
			name: "Valid token",
			setupCookie: func() *http.Cookie {
				token, _ := CreateToken("user1")
				return &http.Cookie{Name: "token", Value: token}
			},
			expectedCode: http.StatusOK,
			expectedBody: `{"message":"Authenticated"}`,
		},
		{
			name: "Expired token",
			setupCookie: func() *http.Cookie {
				claims := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
					"sub": "user1",
					"iss": "pillar-bank",
					"exp": time.Now().Add(-time.Minute).Unix(), // expired 1 minute ago
					"iat": time.Now().Add(-time.Minute).Unix(),
				})
				tokenString, _ := claims.SignedString(secretKey)
				return &http.Cookie{Name: "token", Value: tokenString}
			},
			expectedCode: http.StatusUnauthorized,
			expectedBody: `{"error":"Invalid token"}`,
		},
		{
			name: "Invalid signature",
			setupCookie: func() *http.Cookie {
				wrongKey := []byte("wrong-key")
				claims := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
					"sub": "user1",
					"iss": "pillar-bank",
					"iat": time.Now().Unix(),
				})
				tokenString, _ := claims.SignedString(wrongKey)
				return &http.Cookie{Name: "token", Value: tokenString}
			},
			expectedCode: http.StatusUnauthorized,
			expectedBody: `{"error":"Invalid token"}`,
		},
		{
			name: "Malformed token",
			setupCookie: func() *http.Cookie {
				return &http.Cookie{Name: "token", Value: "malformed-token"}
			},
			expectedCode: http.StatusUnauthorized,
			expectedBody: `{"error":"Invalid token"}`,
		},
		{
			name: "No token",
			setupCookie: func() *http.Cookie {
				return nil
			},
			expectedCode: http.StatusUnauthorized,
			expectedBody: `{"error":"Authentication required"}`,
		},
	}

	// runs tests for each case
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest("GET", "/test", nil)

			if cookie := tt.setupCookie(); cookie != nil {
				req.AddCookie(cookie)
			}

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedCode, w.Code)
			assert.JSONEq(t, tt.expectedBody, w.Body.String())
		})
	}

}
