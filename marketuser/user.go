package marketuser

import (
	"context"
	"errors"
	"github.com/alexedwards/argon2id"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/sirupsen/logrus"
	"net/http"
	"time"
)

type MarketUser struct {
	ID           int    `validate:"-"`
	FirstName    string `binding:"required" validate:"required"`
	LastName     string `binding:"required" validate:"required"`
	NationalCode string `binding:"required" validate:"required"`
	Password     string `binding:"required" validate:"required"`

	LastLogin time.Time
	CreatedAt time.Time

	PasswordHash string
}

type Handlers struct {
	l  *logrus.Logger
	db *pgxpool.Pool
}

func NewHandlers(logger *logrus.Logger, db *pgxpool.Pool) *Handlers {
	return &Handlers{
		l:  logger,
		db: db,
	}
}

func (h *Handlers) GetMarketUser(c *gin.Context) {

}

func (h *Handlers) CreateMarketUser(c *gin.Context) {
	user := MarketUser{}
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	hash, err := argon2id.CreateHash(user.Password, params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	user.PasswordHash = hash

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	tx, err := h.db.Begin(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx, `INSERT INTO market_user (first_name, last_name, national_code, password) VALUES ($1,$2,$3,$4)`,
		user.FirstName,
		user.LastName,
		user.NationalCode,
		user.PasswordHash)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	err = tx.Commit(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
}

var params = &argon2id.Params{
	Memory:      64 * 1024,
	Iterations:  1,
	Parallelism: 2,
	SaltLength:  16,
	KeyLength:   32,
}

type LoginForm struct {
	NationalCode string `binding:"required" validate:"required"`
	Password     string `binding:"required" validate:"required"`
}

var ErrNotFound = errors.New("user not found")
var ErrIncorrectPassword = errors.New("incorrect password")

func (h *Handlers) Login(c *gin.Context) {
	form := LoginForm{}
	if err := c.ShouldBindJSON(&form); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user := &MarketUser{NationalCode: form.NationalCode}
	if err := user.GetPasswordHash(h.db); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": ErrNotFound.Error()})
		return
	}

	match, err := argon2id.ComparePasswordAndHash(form.Password, user.PasswordHash)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if !match {
		c.JSON(http.StatusUnauthorized, gin.H{"error": ErrIncorrectPassword.Error()})
		return
	}
}

func (u *MarketUser) Get(db *pgxpool.Pool) {
}

func (u *MarketUser) GetPasswordHash(db *pgxpool.Pool) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return db.QueryRow(ctx, `SELECT password FROM market_user WHERE national_code=$1`, u.NationalCode).Scan(&u.PasswordHash)

}
