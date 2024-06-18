package db

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	_ "github.com/jackc/pgx/v4/pgxpool"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/redis/go-redis/v9"
)

func NewDB(user string, password string, host string, port string, dbname string, sslmode string) (*sqlx.DB, error) {
	connStr := fmt.Sprintf(
		"user=%s password=%s host=%s port=%v dbname=%s sslmode=%s",
		user,
		password,
		host,
		port,
		dbname,
		sslmode,
	)
	db, err := sqlx.Connect("postgres", connStr)
	if err != nil {
		fmt.Println(err)
		log.Fatal(err)
	}

	err = db.Ping()
	if err != nil {
		fmt.Println("Error connecting to DB: ", err)
	}

	maxOpenConnsStr := os.Getenv("MAX_OPEN_CONNS")
	maxIdleConnsStr := os.Getenv("MAX_IDLE_CONNS")
	maxOpenConns, _ := strconv.ParseInt(maxOpenConnsStr, 10, 64)
	maxIdleConns, _ := strconv.ParseInt(maxIdleConnsStr, 10, 64)
	db.SetMaxOpenConns(int(maxOpenConns))
	db.SetMaxIdleConns(int(maxIdleConns))
	db.SetConnMaxLifetime(5 * time.Minute)

	return db, nil
}

func NewCache() *redis.Client {
	for attempt := 0; attempt < 3; attempt++ {
		cacheClient := redis.NewClient(&redis.Options{
			Addr:     os.Getenv("CACHE_ADDR"),
			Password: os.Getenv("CACHE_PASSWORD"),
			DB:       0,
		})

		status, err := cacheClient.Ping(context.Background()).Result()
		if err != nil {
			fmt.Println("Error connecting to cache. Retrying..", err)
			continue
		}

		fmt.Println("cache status: ", status)

		return cacheClient
	}
	fmt.Println("Failed all attempts to connect to cache client")
	return nil
}
