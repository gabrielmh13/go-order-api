package http

import (
	"net/http"
	"os"
	"time"

	"go-order-api/docs"
	"go-order-api/internal/domain/logger"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

type Server struct {
	router *gin.Engine
	logger logger.Logger
}

type Handler interface {
	SetupRoutes(router *gin.Engine)
}

func NewServer(l logger.Logger, handlers ...Handler) *Server {
	router := gin.Default()

	docs.SwaggerInfo.BasePath = "/"

	router.Use(cors.New(cors.Config{
		AllowWildcard: true,
		AllowOrigins:  []string{"*"},
		AllowMethods:  []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowHeaders:  []string{"Origin", "Content-Type", "Authorization"},
		MaxAge:        time.Hour * 12,
	}))

	for _, handler := range handlers {
		handler.SetupRoutes(router)
	}

	router.GET("/swagger", func(c *gin.Context) {
		c.Redirect(http.StatusTemporaryRedirect, "/swagger/index.html")
	})
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	return &Server{
		router: router,
		logger: l,
	}
}

func (s *Server) Start() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "3333"
	}

	s.logger.Info("Server running on port " + port)
	if err := s.router.Run(":" + port); err != nil {
		s.logger.Error("Error starting HTTP server", err)
		panic(err)
	}
}
