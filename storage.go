package main

import (
	"database/sql"
	"encoding/json"
	"errors"

	"os"

	//"fmt"
	"log"
	"time"

	//"context"

	"github.com/go-redis/redis"

	_ "github.com/lib/pq"
)

type Storage interface {
	Create(album) (album, error)
	Read() ([]album, error)
	ReadOne(id string) (album, error)
	Update(id string, a album) (album, error)
	Delete(id string) error
}

type MemoryStorage struct {
	albums []album
}

func (s MemoryStorage) Create(am album) album {
	s.albums = append(s.albums, am)
	return am
}

func (s MemoryStorage) ReadOne(id string) (album, error) {
	for _, a := range s.albums {
		if a.ID == id {
			return a, nil
		}
	}
	return album{}, errors.New("album not found")
}

func (s MemoryStorage) Read() []album {
	return s.albums
}

func (s MemoryStorage) Update(id string, newAlbum album) (album, error) {
	for i := range s.albums {
		if s.albums[i].ID == id {
			s.albums[i] = newAlbum
			return s.albums[i], nil
		}
	}
	return album{}, errors.New("album not found")
}

func (s MemoryStorage) Delete(id string) error {
	for i, a := range s.albums {
		if a.ID == id {
			s.albums = append(s.albums[:i], s.albums[i+1:]...)
			return nil
		}
	}
	return errors.New("album not found")
}

func NewMemoryStorage() MemoryStorage {
	var albums = []album{
		{ID: "1", Title: "Blue Train", Artist: "John Coltraine", Price: 56.99},
		{ID: "2", Title: "Jeru", Artist: "Gerry Mullingan", Price: 17.99},
		{ID: "3", Title: "Sarah Vaughan and Clifford Brown", Artist: "Sarah Vaughan", Price: 39.99},
	}
	return MemoryStorage{albums: albums}
}

type PostgresStorage struct {
	redisClient *redis.Client
	cacheTTL    time.Duration // Время жизни кеша
	db          *sql.DB
}

func (p PostgresStorage) CreateSchema() error {
	_, err := p.db.Exec("create table if not exists albums (ID char(16) primary key, Title char(128), Artist char(128), Price decimal)")
	return err
}

func NewPostgresStorage(redisClient *redis.Client) PostgresStorage {
	connStr := "user=user dbname=db password=pass sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}
	storage := PostgresStorage{db: db,
		redisClient: redisClient,
	}
	err = storage.CreateSchema()
	if err != nil {
		log.Fatal(err)
	}
	return storage
}

func (p PostgresStorage) Create(am album) (album, error) {
	_, err := p.db.Exec("insert into albums(ID, Title, Artist, Price) values($1, $2, $3, $4)", am.ID, am.Title, am.Artist, am.Price)
	if err != nil {
		return am, err // Возвращаем ошибку, если она произошла
	}
	// Сохранение нового альбома в Redis
	jsonAlbum, err := json.Marshal(am)
	if err != nil {
		log.Println("Error caching album in Redis:", err)
		// Можно вернуть ошибку, если нужно, чтобы Create всегда возвращал ошибку при проблемах с Redis
		// return am, err
	}
	err = p.redisClient.Set("album:"+am.ID, jsonAlbum, p.cacheTTL).Err()
	if err != nil {
		log.Println("Error caching album in Redis:", err)
		// Можно вернуть ошибку, если нужно, чтобы Create всегда возвращал ошибку при проблемах с Redis
		// return am, err
	}

	return am, err
}

func (p PostgresStorage) ReadOne(id string) (album, error) {
	// Попытка получения данных из Redis
	cachedAlbum, err := p.redisClient.Get("album:" + id).Result()
	if err == nil {
		var albums album
		if err := json.Unmarshal([]byte(cachedAlbum), &albums); err == nil {
			return albums, nil
		}
	}
	// Если данных нет в Redis, делаем запрос к PostgreSQL
	var albums album
	row := p.db.QueryRow("select * from albums where id = $1", id)
	err = row.Scan(&albums.ID, &albums.Title, &albums.Artist, &albums.Price)
	if err != nil {
		if err == sql.ErrNoRows {
			return albums, errors.New("not found")
		}
		return albums, err
	}

	// Сохранение полученного альбома в Redis
	jsonAlbum, err := json.Marshal(albums)
	if err != nil {
		log.Println("Error caching album in Redis:", err)
	} else {
		err = p.redisClient.Set("album:"+id, jsonAlbum, p.cacheTTL).Err()
		if err != nil {
			log.Println("Error caching album in Redis:", err)
		}
	}
	return albums, nil
}

func (p PostgresStorage) Read() ([]album, error) {
	// Попытка получения данных из Redis
	cachedAlbums, err := p.redisClient.Get("albums").Result()
	if err == nil {
		var albums []album
		if err := json.Unmarshal([]byte(cachedAlbums), &albums); err == nil {
			return albums, nil
		}
	}
	// Если данных нет в Redis, делаем запрос к PostgreSQL
	var albums []album
	rows, err := p.db.Query("select * from albums")
	if err != nil {
		log.Println("Error querying albums:", err)
		return nil, err //, err
	}
	defer rows.Close()

	for rows.Next() {
		var a album
		if err := rows.Scan(&a.ID, &a.Title, &a.Artist, &a.Price); err != nil {
			return nil, err
		}
		albums = append(albums, a)
	}
	// Сохранение полученных данных в Redis
	jsonAlbums, err := json.Marshal(albums)
	if err != nil {
		return nil, err
	}
	err = p.redisClient.Set("albums", jsonAlbums, p.cacheTTL).Err()
	if err != nil {
		log.Println("Error caching albums in Redis:", err)
	}

	return albums, err //, nil
}

func (p *PostgresStorage) clearCache(id string) error {
	// Очистка кэша всех альбомов
	err := p.redisClient.Del("albums").Err()
	if err != nil {
		log.Println("Error clearing albums cache in Redis:", err)
	}

	// Очистка кэша конкретного альбома
	err = p.redisClient.Del("album:" + id).Err()
	if err != nil {
		log.Println("Error clearing album cache in Redis:", err)
	}
	return err
}

func (p PostgresStorage) Update(id string, a album) (album, error) {
	result, _ := p.db.Exec("update albums set Title=$1, Artist=$2, Price=$3 where id=$4", a.Title, a.Artist, a.Price, id)
	err := handleNotFound(result)
	if err != nil {
		return a, err
	}

	// Очистка кэша после успешного обновления
	err = p.clearCache(id)
	if err != nil {
		return a, err // Передаем ошибку, если произошла ошибка при очистке кэша
	}
	return a, err
}

func (p PostgresStorage) Delete(id string) error {
	result, _ := p.db.Exec("delete from albums where id=$1", id)
	err := handleNotFound(result)
	if err != nil {
		return err
	}
	// Очистка кэша после успешного удаления
	err = p.clearCache(id)
	if err != nil {
		return err // Передаем ошибку, если произошла ошибка при очистке кэша
	}
	return nil

}

func handleNotFound(result sql.Result) error {
	countAffected, _ := result.RowsAffected()
	if countAffected == 0 {
		return errors.New("not found")

	}
	return nil
}

func NewStorage() Storage {
	redisClient := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",            //"some-redis:6379", // адрес вашего Redis-сервера
		Password: os.Getenv("REDIS_PASSWORD"), //"pass",           // пароль, если установлен
		DB:       0,                           // номер базы данных Redis
	})
	return NewPostgresStorage(redisClient)
}
