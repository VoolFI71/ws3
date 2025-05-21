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
// func AuthMiddlewareC() gin.HandlerFunc {
//     return func(c *gin.Context) {
//         tokenString, err := c.Cookie("token")
//         fmt.Println(tokenString)

//         if err != nil || tokenString == "" {
//             c.JSON(http.StatusUnauthorized, gin.H{"error": "Token is missing"})
//             c.Abort()
//             return
//         }

//         token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
//             // Проверка метода подписи
//             if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
//                 return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
//             }
//             return []byte(jwtSecret), nil // Убедитесь, что jwtSecret - это []byte
//         })

//         if err != nil || !token.Valid {
//             c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
//             c.Abort()	
//             return
//         }

//         // Вывод информации о токене
//         claims, ok := token.Claims.(jwt.MapClaims)
//         if ok && token.Valid {
//             fmt.Println("Token claims:", claims) // Выводим содержимое токена
//         } else {
//             fmt.Println("Invalid token claims")
//         }

//         c.Next()
//     }
// }

// func CORSMiddleware() gin.HandlerFunc {
//     return func(c *gin.Context) {
//         // Устанавливаем необходимые заголовки CORS
//         c.Header("Access-Control-Allow-Origin", "*") // Разрешаем все источники
//         c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS") // Разрешаем методы
//         c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Accept") // Разрешаем заголовки

//         // Обрабатываем preflight запросы
//         if c.Request.Method == http.MethodOptions {
//             c.AbortWithStatus(http.StatusNoContent) // Возвращаем статус 204 No Content
//             return
//         }
        
//         c.Next() // Продолжаем обработку запроса
//     }
// }