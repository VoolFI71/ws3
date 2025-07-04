package websocket

import (
	//"database/sql"
	"database/sql"
	"fmt"
	"io"

	//"log"
	"net/http"
	"strings"
	"encoding/base64"

	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"

	//"github.com/go-redis/redis/v8"
	"context"
	"log"

	"github.com/gocql/gocql"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type ChatMessage struct {
	Username  string `json:"username"`
	Message   string `json:"message"`
    CreatedAt int64 `json:"created_at"` 
	Image     string `json:"image"`
	Audio     string `json:"audio"`
    ChatID    int    `json:"chat_id"`
}


var clients = make(map[*websocket.Conn]bool)
var broadcast = make(chan ChatMessage)

func SendMsg()  gin.HandlerFunc { 
	return func (c *gin.Context) {
		conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			fmt.Println("Error while upgrading connection:", err)
			return
		}
		defer conn.Close()

		clients[conn] = true

		for {
			var msg ChatMessage
			err := conn.ReadJSON(&msg)
			if err != nil {
				fmt.Println("Error while reading message:", err)
				delete(clients, conn)
				break
			}
			
			go func(message ChatMessage) {
				broadcast <- message
			}(msg)
		}
	}
}

func SaveMsg(session *gocql.Session) gin.HandlerFunc {
    return func(c *gin.Context) {
        var jwtSecret = []byte("123")

        tokenString := c.GetHeader("Authorization")
        if len(tokenString) > 7 && strings.ToLower(tokenString[:7]) == "bearer " {
            tokenString = tokenString[7:]
        }

        if tokenString == "" {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization token is required"})
            return
        }

        token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
            if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
                return nil, http.ErrNotSupported
            }
            return jwtSecret, nil
        })

        if err != nil || !token.Valid {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
            return
        }

        // Извлечение логина пользователя из токена
        claims, ok := token.Claims.(jwt.MapClaims)
        if !ok || !token.Valid {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims"})
            return
        }
		username, ok := claims["username"].(string)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Username not found in token claims"})
			return
		}

        userIDFloat, ok := claims["user_id"].(float64) 
        fmt.Println(userIDFloat)
        if !ok {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "user_id не найден в данных токена"})
            return
        }
        user_id := int(userIDFloat)
		var messageRequest ChatMessage

        if err := c.BindJSON(&messageRequest); err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
            return
        }
        messageID := gocql.TimeUUID()
        message := messageRequest.Message
        chatID := messageRequest.ChatID
        go func() {
            if session == nil {
                log.Fatalf("Сессия не инициализирована")
            }
            query := session.Query("INSERT INTO messages (chat_id, message_id, username, sender_id, message, created_at) VALUES (?, ?, ?, ?, ?, ?)", chatID, messageID, username, user_id, message, time.Now().Unix())
            if err := query.Exec(); err != nil {
                log.Fatalf("Ошибка при добавлении сообщения в базу данных: %v", err)
                return
            }
            fmt.Println("Сообщение добавлено в базу данных")
        }()
        c.JSON(http.StatusOK, gin.H{"status": "Message saved", "username":  username})
    }
}

func SaveImage(session *gocql.Session) gin.HandlerFunc {
    return func(c *gin.Context) {
        var jwtSecret = []byte("123")

        tokenString := c.GetHeader("Authorization")
        if len(tokenString) > 7 && strings.ToLower(tokenString[:7]) == "bearer " {
            tokenString = tokenString[7:]
        }

        if tokenString == "" {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization token is required"})
            return
        }

        token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
            if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
                return nil, http.ErrNotSupported
            }
            return jwtSecret, nil
        })

        if err != nil || !token.Valid {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
            return
        }

        // Извлечение логина пользователя из токена
        claims, ok := token.Claims.(jwt.MapClaims)
        if !ok || !token.Valid {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims"})
            return
        }
		username, ok := claims["username"].(string)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Username not found in token claims"})
			return
		}

        userIDFloat, ok := claims["user_id"].(float64) 
        fmt.Println(userIDFloat)
        if !ok {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "user_id не найден в данных токена"})
            return
        }
        user_id := int(userIDFloat)

        imageHeader, err := c.FormFile("image")
        if err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": "File upload failed"})
            return
        }
 
        file, err := imageHeader.Open()
        if err != nil {
            fmt.Println("Error opening file:", err)
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to open file"})
            return
        }
        defer file.Close()

        // Читаем содержимое файла
        // image, err := io.ReadAll(file)
        // if err != nil {
        //     fmt.Println("Error reading file:", err)
        //     c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read file"})
        //     return
        // }

        minioClient, err := minio.New("minio:9000", &minio.Options{
            Creds:  credentials.NewStaticV4("123123123", "123123123", ""),
            Secure: false,
        })
        if err != nil {
            fmt.Println("Error creating MinIO client:", err)
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create MinIO client"})
            return
        }
    
        // Имя бакета
        bucketName := "chat-files"
        ctx := context.Background()

        // Создаем бакет, если он не существует 
        exists, err := minioClient.BucketExists(ctx, bucketName)
        if err != nil {
            fmt.Println("Error checking bucket existence:", err)
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check bucket existence"})
            return
        }
        if !exists {
            err = minioClient.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{})
            if err != nil {
                fmt.Println("Error creating bucket:", err)
                c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create bucket"})
                return
            }
        }
        imageUrl := uuid.New().String() // Имя загруженного файла

        // Загружаем изображение в MinIO
        _, err = minioClient.PutObject(ctx, bucketName, imageUrl, file, imageHeader.Size, minio.PutObjectOptions{})
        if err != nil {
            fmt.Println("Error uploading image:", err)
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to upload image"})
            return
        }
        var messageRequest ChatMessage

        if err := c.BindJSON(&messageRequest); err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
            return
        }
        chatID := messageRequest.ChatID
        go func() {
            messageID := gocql.TimeUUID()
            query := session.Query("INSERT INTO messages (chat_id, message_id, username, sender_id, message, image, created_at, audio_data) VALUES (?, ?, ?, ?, ?, ?, ?, ?)", chatID, messageID, username, user_id, "", imageUrl, time.Now().Unix(), nil)
            if err := query.Exec(); err != nil {
                log.Fatalf("Ошибка при добавлении изображения в базу данных %v", err)
            }
        }()

        c.JSON(http.StatusOK, gin.H{"status": "Message saved", "username":  username})
    }
}

func SaveAudio(session *gocql.Session) gin.HandlerFunc {
    return func(c *gin.Context) {
        var jwtSecret = []byte("123")

        tokenString := c.GetHeader("Authorization")
        if len(tokenString) > 7 && strings.ToLower(tokenString[:7]) == "bearer " {
            tokenString = tokenString[7:]
        }

        if tokenString == "" {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization token is required"})
            return
        }

        token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
            if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
                return nil, http.ErrNotSupported
            }
            return jwtSecret, nil
        })

        if err != nil || !token.Valid {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
            return
        }

        // Извлечение логина пользователя из токена
        claims, ok := token.Claims.(jwt.MapClaims)
        if !ok || !token.Valid {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims"})
            return
        }
		username, ok := claims["username"].(string)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Username not found in token claims"})
			return
		}

        userIDFloat, ok := claims["user_id"].(float64) 
        fmt.Println(userIDFloat)
        if !ok {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "user_id не найден в данных токена"})
            return
        }
        user_id := int(userIDFloat)

		audioFile, err := c.FormFile("audio")
		if err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": "No audio file provided"})
            return
        }
		file, err := audioFile.Open()
		if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to open audio file"})
            return
        }
        defer file.Close()


        minioClient, err := minio.New("minio:9000", &minio.Options{
            Creds:  credentials.NewStaticV4("123123123", "123123123", ""),
            Secure: false,
        })
        if err != nil {
            fmt.Println("Error creating MinIO client:", err)
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create MinIO client"})
            return
        }
    
        // Имя бакета
        bucketName := "chat-files"
        ctx := context.Background()

        // Создаем бакет, если он не существует 
        exists, err := minioClient.BucketExists(ctx, bucketName)
        if err != nil {
            fmt.Println("Error checking bucket existence:", err)
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check bucket existence"})
            return
        }
        if !exists {
            err = minioClient.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{})
            if err != nil {
                fmt.Println("Error creating bucket:", err)
                c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create bucket"})
                return
            }
        }
        audioUrl := uuid.New().String() // Имя загруженного файла

        // Загружаем изображение в MinIO
        _, err = minioClient.PutObject(ctx, bucketName, audioUrl, file, audioFile.Size, minio.PutObjectOptions{})
        if err != nil {
            fmt.Println("Error uploading image:", err)
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to upload image"})
            return
        }

        var messageRequest ChatMessage

        if err := c.BindJSON(&messageRequest); err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
            return
        }
        chatID := messageRequest.ChatID
		go func() {
            messageID := gocql.TimeUUID()
            query := session.Query("INSERT INTO messages (chat_id, message_id, username, sender_id, message, image, created_at, audio_data) VALUES (?, ?, ?, ?, ?, ?, ?, ?)", chatID, messageID, username, user_id, "", "", time.Now().Unix(), audioUrl)
            err := query.Exec() // Execute the query and check for errors
            if err != nil {
                fmt.Println(err)
            }
        }()

        c.JSON(http.StatusOK, gin.H{"status": "Message saved", "username":  username})
    }
}

func CreateChat(sessionCassandra *gocql.Session, sessionPG *sql.DB) gin.HandlerFunc {
    return func(c *gin.Context) {
        var jwtSecret = []byte("123")

        tokenString := c.GetHeader("Authorization")
        if len(tokenString) > 7 && strings.ToLower(tokenString[:7]) == "bearer " {
            tokenString = tokenString[7:]
        }

        if tokenString == "" {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization token is required"})
            return
        }

        token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
            if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
                return nil, http.ErrNotSupported
            }
            return jwtSecret, nil
        })

        if err != nil || !token.Valid {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
            return
        }

        claims, ok := token.Claims.(jwt.MapClaims)
        if !ok || !token.Valid {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims"})
            return
        }

        userIDFloat, ok := claims["user_id"].(float64)
        fmt.Println(userIDFloat)
        if !ok {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "user_id не найден в данных токена"})
            return
        }
        user_id:= int(userIDFloat)

        var requestBody struct {
            Name string `json:"namechat"`
        }
        if err := c.ShouldBindJSON(&requestBody); err != nil || requestBody.Name == "" {
            c.JSON(http.StatusBadRequest, gin.H{"error": "Chat name is required"})
            return
        }


        go func() {
            var chatID int
            err := sessionPG.QueryRow("INSERT INTO chats (owner_id, name, lifeupto) VALUES ($1, $2, $3) RETURNING chat_id", user_id, requestBody.Name, time.Now().Unix()+3600).Scan(&chatID)
            if err != nil {
                fmt.Println("Error inserting chat:", err)
                return
            }
            if chatID == 0 {
                fmt.Println("Error: received chat_id is 0")
                return
            }

            _, err = sessionPG.Exec("INSERT INTO chat_members (chat_id, user_id) VALUES ($1, $2)", chatID, user_id)
            if err != nil {
                fmt.Println("Error inserting chat member:", err)
                return
            }
        }()
        c.JSON(http.StatusAccepted, gin.H{"ok": "chat was created"})

    }
}

func GetUserChats(sessionPG *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var jwtSecret = []byte("123")

        tokenString := c.GetHeader("Authorization")
        if len(tokenString) > 7 && strings.ToLower(tokenString[:7]) == "bearer " {
            tokenString = tokenString[7:]
        }

        if tokenString == "" {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization token is required"})
            return
        }

        token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
            if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
                return nil, http.ErrNotSupported
            }
            return jwtSecret, nil
        })

        if err != nil || !token.Valid {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
            return
        }

        claims, ok := token.Claims.(jwt.MapClaims)
        if !ok || !token.Valid {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims"})
            return
        }

        fmt.Println(claims)
        userIDFloat, ok := claims["user_id"].(float64) // Извлекаем как float64
        fmt.Println(userIDFloat)
        if !ok {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "user_id не найден в данных токена"})
            return
        }
        user_id:= int(userIDFloat)
        fmt.Println("user_id:", user_id)

        rows, err := sessionPG.Query(`
            SELECT c.chat_id, c.name 
            FROM chats c
            JOIN chat_members cm ON c.chat_id = cm.chat_id
            WHERE cm.user_id = $1`, user_id)
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось получить чаты"})
            return
        }
        defer rows.Close()

        var chats []struct {
            ChatID int    `json:"chat_id"`
            Name   string `json:"name"`
            // OwnerId int  `json:"owner_id"`

        }
        fmt.Println(err, 111111)

        for rows.Next() {
            var chat struct {
                ChatID int    `json:"chat_id"`
                Name   string `json:"name"`
                // OwnerId int  `json:"owner_id"`
            }

            if err := rows.Scan(&chat.ChatID, &chat.Name); err != nil {
                fmt.Println(err, 333333)
                c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось обработать чат"})
                return
            }
            chats = append(chats, chat)
        }

        if err := rows.Err(); err != nil {
            fmt.Println(err, 44444)
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Произошла ошибка при получении чатов"})
            return
        }

        // Возврат списка чатов в формате JSON
        c.JSON(http.StatusOK, chats)
    }
}
	
type RequestPayload struct {
    ChatID int `json:"chat_id"`
}

func GetMessagesHandler(session *gocql.Session) gin.HandlerFunc {
	return func(c *gin.Context) {
		var payload RequestPayload
		if err := c.ShouldBindJSON(&payload); err != nil {
            fmt.Println("kk")
			c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный запрос"})
			return
		}
        fmt.Println("aaa")
		messages, err := GetLastMessages(session, payload.ChatID)
		if err != nil {
			fmt.Println("Ошибка при получении сообщений:", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось получить сообщения"})
			return
		}

		c.JSON(http.StatusOK, messages)
	}
}

func GetLastMessages(session *gocql.Session, chatID int) ([]ChatMessage, error) {
	iter := session.Query("SELECT username, message, created_at, image, audio_data FROM messages WHERE chat_id=? ORDER BY created_at DESC LIMIT 75", chatID).Iter()
    defer iter.Close()

	var messages []ChatMessage

    minioClient, err := minio.New("minio:9000", &minio.Options{
        Creds:  credentials.NewStaticV4("123123123", "123123123", ""),
        Secure: false,
    })
    if err != nil {
        fmt.Println("Error creating MinIO client:", err)
        return nil, err
    }

    for {
        var msg ChatMessage
        var imageUrl string // Измените на string
        var audioUrl string
        var message string // Измените на string

        if !iter.Scan(&msg.Username, &message, &msg.CreatedAt, &imageUrl, &audioUrl) {
            // Если итерация завершена, выходим из цикла
            break
        }
        msg.Message = message // Присваиваем строку напрямую

        if imageUrl != ""  {
            ctx := context.Background()
            object, err := minioClient.GetObject(ctx, "chat-files", imageUrl, minio.GetObjectOptions{})
            if err != nil {
                fmt.Println("Error getting object:", err)
                return nil, err
            }
            defer object.Close()

            imageData, err := io.ReadAll(object)
            if err != nil {
                fmt.Println("Error reading image data:", err)
                return nil, err
            }
            msg.Image = "data:image/jpeg;base64," + base64.StdEncoding.EncodeToString(imageData)
        }

        if audioUrl != "" {
            ctx := context.Background()
            object, err := minioClient.GetObject(ctx, "chat-files", audioUrl, minio.GetObjectOptions{})
            if err != nil {
                fmt.Println("Error getting object:", err)
                return nil, err
            }
            defer object.Close()
            
            audioData, err := io.ReadAll(object)
            if err != nil {
                fmt.Println("Error reading image data:", err)
                return nil, err
            }
            msg.Audio = "data:audio/wav;base64," + base64.StdEncoding.EncodeToString(audioData)
        }

        messages = append(messages, msg)
    }

    if err := iter.Close(); err != nil {
        fmt.Println("Error closing iterator:", err)
        return nil, err
    }
    fmt.Println("Запрос на вывод сообщений чата ", chatID)
    return messages, nil
}

func HandleMessages() {
	for {
		msg := <-broadcast
		go func(message ChatMessage) {
			for client := range clients {
				err := client.WriteJSON(message)
				if err != nil {
					fmt.Println("Error while writing message:", err)
					client.Close()
					delete(clients, client)
				}
			}
		}(msg)
	}
}