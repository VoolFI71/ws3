package cassandra

import (
    "github.com/gocql/gocql"
    "log"
    "time"

)
type DB struct {
    Session *gocql.Session
}

// NewDB создает новое подключение к Cassandra
func NewDB(host string, keyspace string) *DB {
    cluster := gocql.NewCluster(host) // Используем переданный хост
    cluster.Port = 9042                // Порт по умолчанию
    cluster.Keyspace = keyspace        // Используем переданное ключевое пространство

    // Создаем кластер и подключаемся к Cassandra
    time.Sleep(3 * time.Second) // Задержка перед подключением (можно убрать в реальном коде)

    session, err := cluster.CreateSession()
    if err != nil {
        log.Fatal(err)
    }

    return &DB{Session: session} // Возвращаем указатель на структуру DB
}

// Close закрывает сессию базы данных
func (db *DB) Close() {
    db.Session.Close()
}