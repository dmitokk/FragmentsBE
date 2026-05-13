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
	userFragmentRepo := postgres.NewUserFragmentRepository(db)
	achievementRepo := postgres.NewAchievementRepository(db)

	minioClient, err := minio.NewClient(
		cfg.MinioEndpoint,
		cfg.MinioAccessKey,
		cfg.MinioSecretKey,
		cfg.MinioBucket,
		cfg.MinioUseSSL,
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
		cfg.GoogleAndroidClientID,
	)

	achievementService := service.NewAchievementService(
		achievementRepo,
		userFragmentRepo,
		fragmentRepo,
	)

	fragmentService := service.NewFragmentService(
		fragmentRepo,
		userFragmentRepo,
		achievementService,
		minioClient,
	)

	userService := service.NewUserService(userRepo)

	router := gin.Default()
	http.SetupRoutes(router, authService, fragmentService, userService, achievementService, minioClient)

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
		JWTSecret: getEnv("JWT_SECRET", "your-secret-key-change-in-production"),

		GoogleClientID:       getEnv("GOOGLE_CLIENT_ID", ""),
		GoogleClientSecret:   getEnv("GOOGLE_CLIENT_SECRET", ""),
		GoogleRedirectURL:    getEnv("GOOGLE_REDIRECT_URL", "http://localhost:8080/api/auth/google/callback"),
		GoogleAndroidClientID: getEnv("GOOGLE_ANDROID_CLIENT_ID", ""),
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

	_, err = db.Exec("ALTER TABLE users ADD COLUMN IF NOT EXISTS name VARCHAR(255)")
	if err != nil {
		return fmt.Errorf("failed to add name column: %w", err)
	}

	_, err = db.Exec("ALTER TABLE users ADD COLUMN IF NOT EXISTS avatar_url VARCHAR(255)")
	if err != nil {
		return fmt.Errorf("failed to add avatar_url column: %w", err)
	}

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS user_fragments (
		user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
		fragment_id UUID NOT NULL REFERENCES fragments(id) ON DELETE CASCADE,
		found_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		PRIMARY KEY (user_id, fragment_id)
	)`)
	if err != nil {
		return fmt.Errorf("failed to create user_fragments table: %w", err)
	}

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS achievements (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		code VARCHAR(50) UNIQUE NOT NULL,
		name VARCHAR(255) NOT NULL,
		description TEXT,
		icon_url VARCHAR(255),
		condition_type VARCHAR(50) NOT NULL,
		condition_value INT NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	)`)
	if err != nil {
		return fmt.Errorf("failed to create achievements table: %w", err)
	}

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS user_achievements (
		user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
		achievement_id UUID NOT NULL REFERENCES achievements(id) ON DELETE CASCADE,
		unlocked_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		PRIMARY KEY (user_id, achievement_id)
	)`)
	if err != nil {
		return fmt.Errorf("failed to create user_achievements table: %w", err)
	}

	if err := seedAchievements(db); err != nil {
		return fmt.Errorf("failed to seed achievements: %w", err)
	}

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS fragments (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
		text TEXT,
		geomark GEOMETRY(POINT, 4326),
		sound_url VARCHAR(255),
		photo_urls TEXT[],
		expires_at TIMESTAMP NOT NULL DEFAULT NOW() + INTERVAL '30 days',
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	)`)
	if err != nil {
		return fmt.Errorf("failed to create fragments table: %w", err)
	}

	_, _ = db.Exec("ALTER TABLE fragments ADD COLUMN IF NOT EXISTS expires_at TIMESTAMP NOT NULL DEFAULT NOW() + INTERVAL '30 days'")

	_, _ = db.Exec("CREATE INDEX IF NOT EXISTS idx_fragments_geomark ON fragments USING GIST (geomark)")
	_, _ = db.Exec("CREATE INDEX IF NOT EXISTS idx_fragments_user_id ON fragments (user_id)")
	_, _ = db.Exec("CREATE INDEX IF NOT EXISTS idx_users_email ON users (email)")
	_, _ = db.Exec("CREATE INDEX IF NOT EXISTS idx_users_google_id ON users (google_id)")

	slog.Info("All migrations applied successfully")
	return nil
}

func seedAchievements(db *sql.DB) error {
	achievements := []struct {
		Code           string
		Name           string
		Description    string
		ConditionType  string
		ConditionValue int
	}{
		{"first_found", "Первый шаг", "Найти свой первый осколок", "found_count", 1},
		{"five_found", "Коллекционер", "Найти 5 осколков", "found_count", 5},
		{"ten_found", "Искатель", "Найти 10 осколков", "found_count", 10},
		{"twenty_five_found", "Охотник за воспоминаниями", "Найти 25 осколков", "found_count", 25},
		{"fifty_found", "Легенда", "Найти 50 осколков", "found_count", 50},
		{"with_photo", "Фотограф", "Найти осколок с фотографией", "has_photo", 1},
		{"with_sound", "Аудиофил", "Найти осколок со звуком", "has_sound", 1},
	}

	for _, a := range achievements {
		_, err := db.Exec(`
			INSERT INTO achievements (code, name, description, condition_type, condition_value)
			VALUES ($1, $2, $3, $4, $5)
			ON CONFLICT (code) DO UPDATE
				SET name = EXCLUDED.name,
				    description = EXCLUDED.description,
				    condition_type = EXCLUDED.condition_type,
				    condition_value = EXCLUDED.condition_value
		`, a.Code, a.Name, a.Description, a.ConditionType, a.ConditionValue)
		if err != nil {
			return fmt.Errorf("failed to seed achievement %s: %w", a.Code, err)
		}
	}

	slog.Info("Achievements seeded successfully")
	return nil
}