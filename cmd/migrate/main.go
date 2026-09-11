package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"
	"time"
	"yutagame-backend/infrastructure/database"

	"github.com/go-sql-driver/mysql"
	"github.com/pressly/goose/v3"
	gormmysql "gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func envOrDefault(key, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	return value
}

func mysqlDSN() string {
	config := mysql.NewConfig()
	config.User = os.Getenv("DB_USER")
	config.Passwd = os.Getenv("DB_PASSWORD")
	config.Net = "tcp"
	config.Addr = fmt.Sprintf("%s:%s", envOrDefault("DB_HOST", "127.0.0.1"), envOrDefault("DB_PORT", "3306"))
	config.DBName = os.Getenv("DB_NAME")
	config.ParseTime = true
	config.Loc = time.Local
	config.Params = map[string]string{
		"charset":         "utf8mb4",
		"multiStatements": "true",
	}
	return config.FormatDSN()
}

func openSQLDB() (*sql.DB, error) {
	return sql.Open("mysql", mysqlDSN())
}

func ensureDefaultAdmin() error {
	db, err := gorm.Open(gormmysql.Open(mysqlDSN()), &gorm.Config{})
	if err != nil {
		return err
	}
	return database.EnsureDefaultAdmin(db)
}

func main() {
	command := "up"
	if len(os.Args) > 1 {
		command = strings.TrimSpace(os.Args[1])
	}
	migrationsDir := envOrDefault("MIGRATIONS_DIR", "migrations")

	db, err := openSQLDB()
	if err != nil {
		log.Fatalf("failed to open db: %v", err)
	}
	defer db.Close()

	ctx := context.Background()
	if err := db.PingContext(ctx); err != nil {
		log.Fatalf("failed to connect db: %v", err)
	}
	if err := goose.SetDialect("mysql"); err != nil {
		log.Fatalf("failed to set goose dialect: %v", err)
	}

	switch command {
	case "up":
		if err := goose.UpContext(ctx, db, migrationsDir); err != nil {
			log.Fatalf("migration up failed: %v", err)
		}
		if err := ensureDefaultAdmin(); err != nil {
			log.Fatalf("default admin seed failed: %v", err)
		}
		log.Println("migration up completed")
	case "down":
		if err := goose.DownContext(ctx, db, migrationsDir); err != nil {
			log.Fatalf("migration down failed: %v", err)
		}
		log.Println("migration down completed")
	case "status":
		if err := goose.StatusContext(ctx, db, migrationsDir); err != nil {
			log.Fatalf("migration status failed: %v", err)
		}
	default:
		log.Fatalf("unknown command %q. use up, down, or status", command)
	}
}
