package middleware

import (
	"os"
	"strconv"
	"time"

	"github.com/MShoaei/stock-core/marketuser"
	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/sirupsen/logrus"
)

func NewJWTMiddleware(db *pgxpool.Pool, l *logrus.Logger) *jwt.GinJWTMiddleware {
	// the jwt middleware
	identityKey := "id"
	cfg := &jwt.GinJWTMiddleware{
		Realm:       "test zone",
		Key:         []byte(os.Getenv("SECRET_KEY")),
		Timeout:     time.Hour,
		MaxRefresh:  time.Hour,
		IdentityKey: identityKey,
		PayloadFunc: Payload(identityKey),
		IdentityHandler: func(c *gin.Context) interface{} {
			claims := jwt.ExtractClaims(c)
			return claims[identityKey]
		},
		Authenticator: marketuser.Authenticate(db),

		Authorizator: func(data interface{}, c *gin.Context) bool {
			claimID := int(data.(float64))
			paramID, err := strconv.Atoi(c.Param("id"))
			if err != nil {
				return false
			}
			if paramID != claimID {
				return false
			}

			user := marketuser.MarketUser{ID: claimID}
			if err := user.GetWithID(db); err != nil {
				l.Debugf("authorize failed: %v\n", err)
				return false
			}

			if user.DeletedAt.Valid {
				return false
			}
			return true
		},
		Unauthorized: func(c *gin.Context, code int, message string) {
			c.JSON(code, gin.H{
				"code":    code,
				"message": message,
			})
		},
		TokenLookup: "header: Authorization",

		// TokenHeadName is a string in the header. Default value is "Bearer"
		TokenHeadName: "Bearer",

		// TimeFunc provides the current time. You can override it to use another time value. This is useful for testing or if your server uses a different time zone than your tokens.
		TimeFunc: time.Now,
	}
	authMiddleware, err := jwt.New(cfg)
	if err != nil {
		l.Fatal("JWT Error:" + err.Error())
	}

	return authMiddleware
}

func Payload(identityKey string) func(data interface{}) jwt.MapClaims {
	return func(data interface{}) jwt.MapClaims {
		if v, ok := data.(*marketuser.MarketUser); ok {
			return jwt.MapClaims{
				identityKey: v.ID,
			}
		}
		return jwt.MapClaims{}
	}
}
