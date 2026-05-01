package http

import (
	"github.com/dmitokk/FragmentsBE/internal/application/service"
	"github.com/dmitokk/FragmentsBE/internal/http/handler"
	"github.com/dmitokk/FragmentsBE/internal/http/middleware"
	"github.com/gin-gonic/gin"
)

func SetupRoutes(r *gin.Engine, authService *service.AuthService, fragmentService *service.FragmentService) {
	authHandler := handler.NewAuthHandler(authService)
	fragmentHandler := handler.NewFragmentHandler(fragmentService)

	api := r.Group("/api")
	{
		auth := api.Group("/auth")
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
			auth.GET("/google/url", authHandler.GoogleAuthURL)
			auth.POST("/google", authHandler.GoogleCallback)
		}

		fragments := api.Group("/fragments")
		fragments.Use(middleware.Auth(authService))
		{
			fragments.POST("", fragmentHandler.Create)
			fragments.GET("/:id", fragmentHandler.GetByID)
			fragments.GET("", fragmentHandler.List)
			fragments.PUT("/:id", fragmentHandler.Update)
			fragments.DELETE("/:id", fragmentHandler.Delete)
		}
	}
}