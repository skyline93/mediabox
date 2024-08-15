package auth

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

var jwtSecret = []byte("64lvHM/DzCYuKM97R9HkLaVCCz1T0nrOMn7rBZA1SI0=")

const (
	UserRoleAdmin = "admin"
	UserRoleUser  = "user"
	UserRoleGuest = "guest"
)

func GenerateToken(username, role string) (string, int64) {
	exp := time.Now().Add(time.Hour * 2).Unix()

	claims := jwt.MapClaims{"username": username, "role": role, "exp": exp}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, _ := token.SignedString(jwtSecret)
	return tokenString, exp
}

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Middleware: No token provided"})
			c.Abort()
			return
		}

		if !strings.HasPrefix(authHeader, "Bearer ") {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Middleware: Invalid token format"})
			c.Abort()
			return
		}

		tokenString := authHeader[len("Bearer "):]

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("token method Invalid")
			}
			return jwtSecret, nil
		})

		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Middleware: Invalid token"})
			c.Abort()
			return
		}

		claims, _ := token.Claims.(jwt.MapClaims)
		c.Set("username", claims["username"])
		c.Set("role", claims["role"])
		c.Next()
	}
}

func AdminMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, _ := c.Get("role")

		if role != UserRoleAdmin {
			c.JSON(http.StatusForbidden, gin.H{"error": "Middleware: Forbidden: Insufficient privileges"})
			c.Abort()
			return
		}

		c.Next()
	}
}
