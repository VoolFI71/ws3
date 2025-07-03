package middleware

import (
    //"fmt"
    _ "github.com/jackc/pgx/v4/stdlib"
    "github.com/gin-gonic/gin"
    "net/http"
    "github.com/golang-jwt/jwt/v4"
	"strings"
    //"github.com/gin-contrib/sessions"
    //"github.com/gin-contrib/sessions/cookie"
)

var jwtSecret = []byte("123")

func AuthMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        authHeader := c.GetHeader("Authorization")
        if authHeader == "" {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header is missing"})
            c.Abort()
            return
        }

        // Проверка формата заголовка
        parts := strings.Split(authHeader, " ")
        if len(parts) != 2 || parts[0] != "Bearer" {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization format"})
            c.Abort()
            return
        }

        tokenString := parts[1]

        token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
            return jwtSecret, nil
        })

        if err != nil || !token.Valid {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
            c.Abort()	
            return
        }

        c.Next()
    }
}
