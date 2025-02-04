package postgres

import (
	"embed"
	"fmt"
	"log"
	"time"

	"github.com/dapplux/twitter-haiku-bot/config"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	pgsql "gorm.io/driver/postgres"
	"gorm.io/gorm"
)

//go:embed migrations/*.sql
var migrationFiles embed.FS

// Database wraps GORM DB
type Database struct {
	DB *gorm.DB
}

// New initializes PostgreSQL and runs migrations (up by default).
// If you want to run down migrations, pass direction="down" or a partial version.
func New(cfg config.Config, direction string, version int) (*Database, error) {
	db, err := connectDB(cfg)
	if err != nil {
		return nil, err
	}

	if err := runMigrations(db, direction, version); err != nil {
		return nil, err
	}

	return &Database{DB: db}, nil
}

// connectDB initializes the PostgreSQL connection with GORM
func connectDB(cfg config.Config) (*gorm.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%d sslmode=%s TimeZone=UTC",
		cfg.DB.Host, cfg.DB.User, cfg.DB.Password, cfg.DB.Name, cfg.DB.Port, cfg.DB.SSLMode,
	)

	db, err := gorm.Open(pgsql.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %v", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get sql.DB instance: %v", err)
	}

	// Optional connection pool configuration
	sqlDB.SetMaxOpenConns(25)
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetConnMaxLifetime(5 * time.Minute)

	// Validate connection
	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("database connection test failed: %v", err)
	}

	return db, nil
}

// runMigrations applies up or down migrations, possibly to a specific version.
func runMigrations(db *gorm.DB, direction string, version int) error {
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("failed to get sql.DB instance: %v", err)
	}

	driver, err := postgres.WithInstance(sqlDB, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("could not create migration driver: %v", err)
	}

	d, err := iofs.New(migrationFiles, "migrations")
	if err != nil {
		return fmt.Errorf("could not load embedded migration files: %v", err)
	}

	m, err := migrate.NewWithInstance("iofs", d, "postgres", driver)
	if err != nil {
		return fmt.Errorf("could not initialize migration: %v", err)
	}

	// Decide which direction to run (default: up)
	switch direction {
	case "down":
		if version > 0 {
			err = m.Steps(-1 * int(version)) // Revert specific number of steps
		} else {
			err = m.Down() // Revert all migrations
		}
	default:
		// Up migrations
		if version > 0 {
			err = m.Migrate(uint(version)) // Migrate up to a specific version
		} else {
			err = m.Up() // Migrate all the way up
		}
	}

	if err == migrate.ErrNoChange {
		log.Println("No new migrations to apply.")
		return nil
	} else if err != nil {
		return fmt.Errorf("migration failed: %v", err)
	}

	log.Println("Migrations applied successfully.")
	return nil
}

// ListEmbeddedFiles prints the embedded migration files for debugging
func ListEmbeddedFiles() {
	files, err := migrationFiles.ReadDir("migrations")
	if err != nil {
		log.Println("Failed to read embedded migrations:", err)
		return
	}

	log.Println("Embedded migration files:")
	for _, file := range files {
		log.Println("-", file.Name())
	}
}
