package server

import (
	"github.com/MShoaei/stock-core/marketuser"
	"github.com/MShoaei/stock-core/middleware"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/sirupsen/logrus"
)

func New(logger *logrus.Logger, db *pgxpool.Pool) *gin.Engine {
	e := gin.New()
	e.Use(gin.Logger(), gin.Recovery())

	authMiddleware := middleware.NewJWTMiddleware(db, logger)

	e.POST("/login", authMiddleware.LoginHandler)

	{
		authGroup := e.Group("/auth")
		authGroup.GET("/refresh-token", authMiddleware.RefreshHandler)
		authGroup.GET("/logout", authMiddleware.LogoutHandler)
	}

	{
		mu := marketuser.NewHandlers(logger, db)
		userGroup := e.Group("/users")
		userGroup.POST("/", mu.CreateMarketUser)

		userGroup.Use(authMiddleware.MiddlewareFunc())
		userGroup.GET("/:id", mu.GetMarketUser)
		userGroup.PATCH("/:id", mu.UpdateMarketUser)
		userGroup.DELETE("/:id", mu.DeleteMarketUser)
	}

	return e
}
