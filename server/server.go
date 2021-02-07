package server

import (
	"github.com/MShoaei/stock-core/marketuser"
	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/sirupsen/logrus"
)

func New(log *logrus.Logger, db *pgxpool.Pool, authMiddleware *jwt.GinJWTMiddleware) *gin.Engine {
	e := gin.Default()

	e.POST("/login", authMiddleware.LoginHandler)

	{
		authGroup := e.Group("/auth")
		authGroup.GET("/refresh-token", authMiddleware.RefreshHandler)
		authGroup.GET("/logout", authMiddleware.LogoutHandler)
	}

	{
		mu := marketuser.NewHandlers(log, db)
		userGroup := e.Group("/users")
		userGroup.POST("/", mu.CreateMarketUser)
		userGroup.Use(authMiddleware.MiddlewareFunc())
		userGroup.GET("/", mu.GetMarketUser)
		//userGroup.PATCH()
		//userGroup.DELETE()
	}

	return e
}
