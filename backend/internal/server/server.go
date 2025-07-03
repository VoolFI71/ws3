package server

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
)

var jwtSecret = []byte("123")

func GetUsername(c *gin.Context) string {
        tokenString := c.GetHeader("Authorization")
        
        // Удаляем "Bearer " из токена, если он присутствует
        if len(tokenString) > 7 && tokenString[:7] == "Bearer " {
            tokenString = tokenString[7:]
        }

        if tokenString == "" {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization token is required"})
            return ""
        }

        token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
            if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
                return nil, http.ErrNotSupported
            }
            return jwtSecret, nil
        })

        if err != nil || !token.Valid {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
            return ""
        }

        claims, ok := token.Claims.(jwt.MapClaims)
        if !ok || !token.Valid {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims"})
            return ""
        }

        username, ok := claims["username"].(string)
        if !ok {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "Username not found in token"})
            return ""
        }
        return username
    }

	
func GetUserId(c *gin.Context) int {
	tokenString := c.GetHeader("Authorization")
	
	// Удаляем "Bearer " из токена, если он присутствует
	if len(tokenString) > 7 && tokenString[:7] == "Bearer " {
		tokenString = tokenString[7:]
	}

	if tokenString == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization token is required"})
		return 0
	}

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, http.ErrNotSupported
		}
		return jwtSecret, nil
	})

	if err != nil || !token.Valid {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
		return 0
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims"})
		return 0
	}

	user_id, ok := claims["user_id"].(int)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user_id not found in token"})
		return 0
	}
	return user_id
}