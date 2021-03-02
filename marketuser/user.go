package marketuser

import (
	"context"
	"database/sql"
	"errors"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"net/http"
	"time"

	"github.com/alexedwards/argon2id"
	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

var params = &argon2id.Params{
	Memory:      64 * 1024,
	Iterations:  1,
	Parallelism: 2,
	SaltLength:  16,
	KeyLength:   32,
}

type MarketUser struct {
	ID           int    `db:"id"`
	FirstName    string `binding:"required" db:"first_name"`
	LastName     string `binding:"required" db:"last_name"`
	NationalCode string `binding:"required" db:"national_code"`
	Password     string `binding:"required" db:"password"`

	LastLogin sql.NullTime `json:"-" db:"last_login"`
	CreatedAt time.Time    `json:"-" db:"created_at"`
	DeletedAt sql.NullTime `json:"-" db:"deleted_at"`
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

func (h *Handlers) CreateMarketUser(c *gin.Context) {
	user := MarketUser{}
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	if err := user.GetWithNationalCode(h.db); err != nil && !errors.Is(err, pgx.ErrNoRows) {
		h.l.Errorf("database error: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to retrieve user",
		})
		return
	}
	if user.ID != 0 {
		c.JSON(http.StatusConflict, gin.H{
			"error": "user already exists",
		})
		return
	}

	if err := user.Create(h.db); err != nil {
		h.l.Errorf("database error: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to create user",
		})
		return
	}
	c.JSON(http.StatusCreated, gin.H{
		"message": "created user",
	})
}

func (h *Handlers) GetMarketUser(c *gin.Context) {
	claims := jwt.ExtractClaims(c)
	user := MarketUser{ID: int(claims["id"].(float64))}

	if err := user.GetWithID(h.db); err != nil {
		h.l.Errorf("database error: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to retrieve user",
		})
		return
	}
	if user.DeletedAt.Valid {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "user does not exist",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user": gin.H{
			"FirstName":    user.FirstName,
			"LastName":     user.LastName,
			"NationalCode": user.NationalCode,
		},
	})
}

func (u *MarketUser) GetWithID(db *pgxpool.Pool) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	const userQuery = `SELECT first_name, last_name, national_code, password, last_login, created_at, deleted_at FROM market_user WHERE id=$1`
	return db.QueryRow(ctx, userQuery, u.ID).Scan(
		&u.FirstName, &u.LastName, &u.NationalCode, &u.Password, &u.LastLogin, &u.CreatedAt, &u.DeletedAt)
}

func (u *MarketUser) GetWithNationalCode(db *pgxpool.Pool) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	const userQuery = `SELECT id, first_name, last_name, password, last_login, created_at, deleted_at FROM market_user WHERE national_code=$1`
	return db.QueryRow(ctx, userQuery, u.NationalCode).Scan(
		&u.ID, &u.FirstName, &u.LastName, &u.Password, &u.LastLogin, &u.CreatedAt, &u.DeletedAt)
}

func (h *Handlers) UpdateMarketUser(c *gin.Context) {
	data := struct {
		Password string `binding:"required" validate:"required" db:"password"`
	}{}
	if err := c.ShouldBindJSON(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	claims := jwt.ExtractClaims(c)
	user := MarketUser{ID: int(claims["id"].(float64)), Password: data.Password}

	if err := user.Update(h.db); err != nil {
		h.l.Debugf("database error: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to update user",
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{})
}

func (h *Handlers) DeleteMarketUser(c *gin.Context) {
	claims := jwt.ExtractClaims(c)
	user := MarketUser{ID: int(claims["id"].(float64))}

	if err := user.Delete(h.db); err != nil {
		h.l.Debugf("database error: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to delete user",
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{})
}

// Create creates a new user and adds it to the database. Any error returned is related to database.
func (u *MarketUser) Create(db *pgxpool.Pool) error {
	hash, err := argon2id.CreateHash(u.Password, params)
	if err != nil {
		panic(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	tx, err := db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	const insertUserQuery = `INSERT INTO market_user (first_name, last_name, national_code, password) VALUES ($1,$2,$3,$4)`
	_, err = tx.Exec(ctx, insertUserQuery,
		u.FirstName,
		u.LastName,
		u.NationalCode,
		hash)
	if err != nil {
		return err
	}
	err = tx.Commit(ctx)
	if err != nil {
		return err
	}
	return nil
}

// Update updates the user info. Only changing the password is supported. Any error returned is related to database.
func (u *MarketUser) Update(db *pgxpool.Pool) error {
	hash, err := argon2id.CreateHash(u.Password, params)
	if err != nil {
		panic(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	const updatePasswordQuery = `UPDATE market_user SET password = $1 WHERE id=$2 AND deleted_at IS NULL`
	commandTag, err := db.Exec(ctx, updatePasswordQuery, hash, u.ID)
	if err != nil {
		return err
	}

	// since the user is authorized before any action this can not fail.
	if commandTag.RowsAffected() != 1 {
		panic("could not update user password")
	}
	return nil
}

func (u *MarketUser) Delete(db *pgxpool.Pool) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	const deleteUserQuery = `UPDATE market_user SET deleted_at = NOW() WHERE id=$1 AND deleted_at IS NULL`
	commandTag, err := db.Exec(ctx, deleteUserQuery, u.ID)
	if err != nil {
		return err
	}

	// since the user is authorized before any action this can not fail.
	if commandTag.RowsAffected() != 1 {
		panic("could not delete user")
	}
	return nil
}

type LoginForm struct {
	NationalCode string `binding:"required" validate:"required"`
	Password     string `binding:"required" validate:"required"`
}

func Authenticate(db *pgxpool.Pool) func(c *gin.Context) (interface{}, error) {
	return func(c *gin.Context) (interface{}, error) {
		form := LoginForm{}
		if err := c.ShouldBindJSON(&form); err != nil {
			return nil, err
		}

		user := &MarketUser{NationalCode: form.NationalCode}
		if err := user.GetWithNationalCode(db); err != nil {
			return nil, err //TODO: is it a security problem?
		}
		if user.DeletedAt.Valid {
			return nil, jwt.ErrFailedAuthentication
		}
		logrus.Println(user)
		match, err := argon2id.ComparePasswordAndHash(form.Password, user.Password)
		if err != nil {
			return nil, err
		}
		if !match {
			return nil, jwt.ErrFailedAuthentication
		}
		return user, nil
	}
}
