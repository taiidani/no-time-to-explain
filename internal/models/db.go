package models

import (
	"context"
	"database/sql"
	"os"

	internalDB "github.com/taiidani/no-time-to-explain/internal/db"
)

var db *sql.DB

func InitDB(ctx context.Context) error {
	client, err := internalDB.New(ctx, os.Getenv("DATABASE_URL"))
	db = client
	return err
}
