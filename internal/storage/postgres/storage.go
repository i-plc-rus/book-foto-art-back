package postgres

import (
	"context"
	"fmt"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Storage struct {
	DB         *pgxpool.Pool
	User       *UserStorage
	Collection *CollectionStorage
	Upload     *UploadStorage
}

func InitDB() *Storage {
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s",
		os.Getenv("DB_USER"), os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_HOST"), os.Getenv("DB_PORT"),
		os.Getenv("DB_NAME"))

	dbpool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		panic(err)
	}

	return &Storage{
		DB:         dbpool,
		User:       &UserStorage{DB: dbpool},
		Collection: &CollectionStorage{DB: dbpool},
		Upload:     &UploadStorage{DB: dbpool},
	}
}
