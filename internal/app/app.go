package app

import (
	"database/sql"
	"fmt"
	"log/slog"
	"os"

	"github.com/dmitokk/FragmentsBE/internal/application/service"
	"github.com/dmitokk/FragmentsBE/internal/infrastructure/persistence/postgres"
	"github.com/dmitokk/FragmentsBE/internal/infrastructure/storage/minio"
	"github.com/dmitokk/FragmentsBE/internal/http"
	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
)

type App struct {
	cfg *Config

	router *gin.Engine

	db *sql.DB
}

func (a *App) Run() error {
	slog.Info("Starting server", "port", a.cfg.HTTPPort)
	return a.router.Run(":" + a.cfg.HTTPPort)
}

func New() (*App, error) {
	cfg := loadConfig()

	db, err := sql.Open("postgres", cfg.DBUrl)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	if err := runMigrations(db); err != nil {
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	userRepo := postgres.NewUserRepository(db)
	fragmentRepo := postgres.NewFragmentRepository(db)

	minioClient, err := minio.NewClient(
		cfg.MinioEndpoint,
		cfg.MinioAccessKey,
		cfg.MinioSecretKey,
		cfg.MinioBucket,
		cfg.MinioUseSSL,
		cfg.MinioPublicURL,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create minio client: %w", err)
	}

	authService := service.NewAuthService(
		userRepo,
		cfg.JWTSecret,
		cfg.GoogleClientID,
		cfg.GoogleClientSecret,
		cfg.GoogleRedirectURL,
	)

	fragmentService := service.NewFragmentService(fragmentRepo, minioClient)

	router := gin.Default()
	http.SetupRoutes(router, authService, fragmentService)

	slog.Info("Application initialized")

	return &App{
		cfg:    cfg,
		router: router,
		db:     db,
	}, nil
}

func loadConfig() *Config {
	return &Config{
		HTTPPort: getEnv("HTTP_PORT", "8080"),

		DBUrl: getEnv("DB_URL", "postgres://fragments:fragments@localhost:5432/fragments?sslmode=disable"),

		MinioEndpoint:   getEnv("MINIO_ENDPOINT", "localhost:9000"),
		MinioAccessKey:  getEnv("MINIO_ACCESS_KEY", "minioadmin"),
		MinioSecretKey:  getEnv("MINIO_SECRET_KEY", "minioadmin"),
		MinioBucket:     getEnv("MINIO_BUCKET", "fragments"),
		MinioUseSSL:     getEnv("MINIO_USE_SSL", "false") == "true",
		MinioPublicURL:  getEnv("MINIO_PUBLIC_URL", "http://localhost:9000"),

		JWTSecret: getEnv("JWT_SECRET", "your-secret-key-change-in-production"),

		GoogleClientID:     getEnv("GOOGLE_CLIENT_ID", ""),
		GoogleClientSecret: getEnv("GOOGLE_CLIENT_SECRET", ""),
		GoogleRedirectURL:  getEnv("GOOGLE_REDIRECT_URL", "http://localhost:8080/api/auth/google/callback"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func runMigrations(db *sql.DB) error {
	var exists int
	err := db.QueryRow("SELECT 1 FROM pg_extension WHERE extname = 'postgis'").Scan(&exists)
	if err == sql.ErrNoRows {
		if _, err := db.Exec("CREATE EXTENSION IF NOT EXISTS postgis"); err != nil {
			return fmt.Errorf("failed to create postgis extension: %w", err)
		}
	}

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS users (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		email VARCHAR(255) UNIQUE NOT NULL,
		password_hash VARCHAR(255),
		google_id VARCHAR(255),
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	)`)
	if err != nil {
		return fmt.Errorf("failed to create users table: %w", err)
	}

	_, _ = db.Exec("ALTER TABLE users DROP CONSTRAINT IF EXISTS users_google_id_key")

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS fragments (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
		text TEXT,
		geomark GEOMETRY(POINT, 4326),
		sound_url VARCHAR(255),
		photo_urls TEXT[],
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	)`)
	if err != nil {
		return fmt.Errorf("failed to create fragments table: %w", err)
	}

	_, _ = db.Exec("CREATE INDEX IF NOT EXISTS idx_fragments_geomark ON fragments USING GIST (geomark)")
	_, _ = db.Exec("CREATE INDEX IF NOT EXISTS idx_fragments_user_id ON fragments (user_id)")
	_, _ = db.Exec("CREATE INDEX IF NOT EXISTS idx_users_email ON users (email)")
	_, _ = db.Exec("CREATE INDEX IF NOT EXISTS idx_users_google_id ON users (google_id)")

	slog.Info("All migrations applied successfully")
	return nil
}