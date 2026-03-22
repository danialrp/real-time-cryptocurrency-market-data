package database

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"
	
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	_ "github.com/lib/pq"
)

func CreateOrConnectDatabase() {
	dbUser := os.Getenv("DB_USERNAME")
	dbPass := os.Getenv("DB_PASSWORD")
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbName := os.Getenv("DB_DATABASE")
	dbSslMode := os.Getenv("DB_SSLMODE")

	if dbSslMode == "" {
		dbSslMode = "disable"
	}

	checkDSN := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=postgres sslmode=%s",
		dbHost, dbPort, dbUser, dbPass, dbSslMode)

	db, err := sql.Open("postgres", checkDSN)
	if err != nil {
		log.Fatalf("❌ Failed to connect to Postgres: %v", err)
	}
	defer db.Close()

	var exists bool
	err = db.QueryRow("SELECT EXISTS (SELECT 1 FROM pg_database WHERE datname = $1)", dbName).Scan(&exists)
	if err != nil {
		log.Fatalf("❌ Failed to check database existence: %v", err)
	}

	if exists {
		log.Printf("✅ Database '%s' checked and Ok.\n", dbName)
	} else {
		_, err := db.Exec(fmt.Sprintf("CREATE DATABASE \"%s\"", dbName))
		if err != nil {
			log.Fatalf("❌ Failed to create database '%s': %v", dbName, err)
		}
		log.Printf("📦 Created database '%s' successfully.\n", dbName)
	}
}

func NewPostgresConnection() (*gorm.DB, error) {
	dbUser := os.Getenv("DB_USERNAME")
	dbPass := os.Getenv("DB_PASSWORD")
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbName := os.Getenv("DB_DATABASE")
	dbSslMode := os.Getenv("DB_SSLMODE")

	if dbSslMode == "" {
		dbSslMode = "disable"
	}

	log.Println("🔧 Preparing DSN for Postgres connection...")

	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s TimeZone=Asia/Tehran",
		dbHost, dbPort, dbUser, dbPass, dbName, dbSslMode,
	)

	log.Println("🟡 Connecting to PostgreSQL using GORM...")

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Printf("❌ Failed to connect to Postgres: %v\n", err)
		return nil, err
	}
	log.Println("✅ GORM connected to PostgreSQL successfully.")

	sqlDB, err := db.DB()
	if err != nil {
		log.Printf("❌ Failed to get sql.DB: %v\n", err)
		return nil, err
	}

	sqlDB.SetMaxOpenConns(10)
	sqlDB.SetConnMaxLifetime(time.Hour)
	log.Println("🔧 Connection pool settings applied.")

	if err := MigrateDatabaseTables(db); err != nil {
		log.Fatalf("❌ Migration failed: %v", err)
		return nil, err
	}

	return db, nil
}

