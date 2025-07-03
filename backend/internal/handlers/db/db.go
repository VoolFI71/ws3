package db

import (
    "database/sql"
    "fmt"
    //"log"

)

var database *sql.DB

func Connect() error {
    var err error
    database, err = sql.Open("postgres", "postgresql://postgres:1234@db:5432/go?sslmode=disable")
    if err != nil {
        return err
    }

    if err := database.Ping(); err != nil {
        return err
    }

    _, err = database.Exec(`
        
            CREATE TABLE IF NOT EXISTS users (
                user_id SERIAL PRIMARY KEY,          -- Уникальный идентификатор для каждого пользователя
                username VARCHAR(50) UNIQUE,
                password VARCHAR(100) NOT NULL,
                email VARCHAR(50) UNIQUE NOT NULL
            );
            
            CREATE TABLE IF NOT EXISTS chats (
                chat_id BIGSERIAL PRIMARY KEY,  -- Автоинкремент для уникального идентификатора чата
                name TEXT NOT NULL, -- НАЗВАНИЕ ЧАТА
                owner_id INT NOT NULL,          -- Владелец чата
                lifeupto INT,
                FOREIGN KEY (owner_id) REFERENCES users(user_id) ON DELETE CASCADE -- Обеспечивает целостность ссылок
            );
            
            CREATE TABLE IF NOT EXISTS chat_members (
                chat_id INT NOT NULL,                 -- Внешний ключ на таблицу чатов
                user_id INT NOT NULL,                 -- Внешний ключ на таблицу пользователей
                PRIMARY KEY (chat_id, user_id),      -- Композитный первичный ключ
                FOREIGN KEY (chat_id) REFERENCES chats(chat_id) ON DELETE CASCADE, -- Обеспечивает целостность ссылок
                FOREIGN KEY (user_id) REFERENCES users(user_id) ON DELETE CASCADE  -- Обеспечивает целостность ссылок
            );
       `)

    if err != nil {
        return fmt.Errorf("ошибка при создании таблиц: %w", err)
    }

    return nil
}

func GetDB() *sql.DB {
    return database
}

func Close() {
    if database != nil {
        database.Close()
    }
}
