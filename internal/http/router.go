package http

import (
	"github.com/dmitokk/FragmentsBE/internal/application/service"
	"github.com/dmitokk/FragmentsBE/internal/http/handler"
	"github.com/dmitokk/FragmentsBE/internal/http/middleware"
	"github.com/dmitokk/FragmentsBE/internal/infrastructure/storage/minio"
	"github.com/gin-gonic/gin"
)

func SetupRoutes(
	r *gin.Engine,
	authService *service.AuthService,
	fragmentService *service.FragmentService,
	userService *service.UserService,
	achievementService *service.AchievementService,
	minioClient *minio.Client,
) {
	authHandler := handler.NewAuthHandler(authService)
	fragmentHandler := handler.NewFragmentHandler(fragmentService)
	userHandler := handler.NewUserHandler(userService, minioClient)
	achievementHandler := handler.NewAchievementHandler(achievementService)
	fileHandler := handler.NewFileHandler(minioClient)

	api := r.Group("/api")
	{
		auth := api.Group("/auth")
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
			auth.GET("/google/url", authHandler.GoogleAuthURL)
			auth.POST("/google", authHandler.GoogleCallback)
			auth.POST("/google/android", authHandler.GoogleAndroidAuth)
		}

		users := api.Group("/users")
		users.Use(middleware.Auth(authService))
		{
			users.GET("/profile", userHandler.GetProfile)
			users.GET("/:id", userHandler.GetUserByID)
			users.PUT("/profile", userHandler.UpdateProfile)
		}

		fragments := api.Group("/fragments")
		fragments.Use(middleware.Auth(authService))
		{
			fragments.POST("", fragmentHandler.Create)
			fragments.GET("", fragmentHandler.List)
			fragments.GET("/found", fragmentHandler.GetFound)
			fragments.POST("/:id/found", fragmentHandler.MarkFound)
			fragments.GET("/:id", fragmentHandler.GetByID)
		}

		files := api.Group("/files")
		files.Use(middleware.Auth(authService))
		{
			files.GET("/*filepath", fileHandler.ServeFile)
		}

		achievements := api.Group("/achievements")
		achievements.Use(middleware.Auth(authService))
		{
			achievements.GET("", achievementHandler.GetAll)
			achievements.GET("/mine", achievementHandler.GetMine)
		}
	}
}