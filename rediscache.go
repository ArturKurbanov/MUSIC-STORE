package main

import (
	//"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/go-redis/redis"
	//"github.com/redis/go-redis/v9"
)

var redisClient *redis.Client

func initRedis() {
	redisClient = redis.NewClient(&redis.Options{
		Addr:     "localhost:6379", // адрес вашего Redis-сервера
		Password: "pass",           // пароль, если установлен
		DB:       0,                // номер базы данных Redis
	})

	// Проверка соединения с Redis
	pong, err := redisClient.Ping().Result()
	if err != nil {
		panic(err)
	}
	fmt.Println("Connected to Redis:", pong)
	//return NewPostgresStorage(redisClient) // Передайте Redis-клиент
}

type RedisCache struct {
	*PostgresStorage
	redisClient *redis.Client
	cacheTTL    time.Duration // Время жизни кеша
}

func NewRedisStorage(db *sql.DB, redisClient *redis.Client) *RedisCache {
	return &RedisCache{
		PostgresStorage: &PostgresStorage{ // Создаем экземпляр PostgresStorage
			db: db,
		},
		redisClient: redisClient,
	}
}

// func (rc *RedisCache) Read() ([]album, error) {
// 	// Попытка получения данных из Redis
// 	cachedAlbums, err := rc.redisClient.Get("albums").Result()
// 	if err == redis.Nil {
// 		// Данные не найдены в Redis, получаем из базы данных
// 		albums := rc.PostgresStorage.Read() // Передаем функцию

// 		// Кэшируем полученные данные в Redis
// 		cachedData, err := json.Marshal(albums)
// 		if err != nil {
// 			return nil, err
// 		}

// 		err = rc.redisClient.Set("albums", cachedData, rc.cacheTTL).Err()
// 		if err != nil {
// 			return nil, err
// 		}

// 		return albums, nil
// 	} else if err != nil {
// 		// Ошибка Redis, логируем и получаем данные из базы данных
// 		fmt.Println("Ошибка Redis:", err)
// 		return rc.PostgresStorage.Read(), nil
// 	}
// 	// Данные найдены в Redis, десериализуем и возвращаем
// 	var albums []album
// 	if err := json.Unmarshal([]byte(cachedAlbums), &albums); err != nil {
// 		return nil, err
// 	}
// 	return albums, nil
// }

// Функция для получения данных из базы данных
//func (rc *RedisCache) getFromDatabase() ([]album, error) {
//	return rc.PostgresStorage.Read(rc.getFromDatabase), nil
//}
