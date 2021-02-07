package main

import (
	"os"
	"time"

	"github.com/MShoaei/stock-core/marketuser"
	"github.com/alexedwards/argon2id"
	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v4/pgxpool"
)

func NewJWTMiddleware(db *pgxpool.Pool) *jwt.GinJWTMiddleware {
	// the jwt middleware
	identityKey := "id"
	cfg := &jwt.GinJWTMiddleware{
		Realm:       "test zone",
		Key:         []byte(os.Getenv("SECRET_KEY")),
		Timeout:     time.Hour,
		MaxRefresh:  time.Hour,
		IdentityKey: identityKey,
		PayloadFunc: func(data interface{}) jwt.MapClaims {
			if v, ok := data.(*marketuser.MarketUser); ok {
				return jwt.MapClaims{
					identityKey: v.ID,
				}
			}
			return jwt.MapClaims{}
		},
		IdentityHandler: func(c *gin.Context) interface{} {
			claims := jwt.ExtractClaims(c)
			return &marketuser.MarketUser{
				ID: int(claims[identityKey].(float64)),
			}
		},
		Authenticator: func(c *gin.Context) (interface{}, error) {
			form := marketuser.LoginForm{}
			if err := c.ShouldBindJSON(&form); err != nil {
				return nil, err
			}

			user := &marketuser.MarketUser{NationalCode: form.NationalCode}
			if err := user.GetPasswordHash(db); err != nil {
				return nil, err
			}

			match, err := argon2id.ComparePasswordAndHash(form.Password, user.PasswordHash)
			if err != nil {
				return nil, err
			}
			if !match {
				return nil, jwt.ErrFailedAuthentication
			}

			return user, nil
		},
		Authorizator: func(data interface{}, c *gin.Context) bool {
			return true
		},
		Unauthorized: func(c *gin.Context, code int, message string) {
			c.JSON(code, gin.H{
				"code":    code,
				"message": message,
			})
		},
		// TokenLookup is a string in the form of "<source>:<name>" that is used
		// to extract token from the request.
		// Optional. Default value "header:Authorization".
		// Possible values:
		// - "header:<name>"
		// - "query:<name>"
		// - "cookie:<name>"
		// - "param:<name>"
		TokenLookup: "header: Authorization, query: token, cookie: jwt",
		// TokenLookup: "query:token",
		// TokenLookup: "cookie:token",

		// TokenHeadName is a string in the header. Default value is "Bearer"
		TokenHeadName: "Bearer",

		// TimeFunc provides the current time. You can override it to use another time value. This is useful for testing or if your server uses a different time zone than your tokens.
		TimeFunc: time.Now,
	}
	authMiddleware, err := jwt.New(cfg)
	if err != nil {
		log.Fatal("JWT Error:" + err.Error())
	}

	return authMiddleware
}
