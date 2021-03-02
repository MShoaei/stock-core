package marketuser_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/MShoaei/stock-core/server"
	"github.com/bxcodec/faker/v3"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v4/pgxpool"
	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/jmoiron/sqlx"
	migrate "github.com/rubenv/sql-migrate"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

var db *pgxpool.Pool
var api *gin.Engine
var logger = logrus.New()

func doMigrations(connString string) {
	migrations := &migrate.FileMigrationSource{
		Dir: "../migrations",
	}
	db := sqlx.MustConnect("pgx", connString)
	_, err := migrate.Exec(db.DB, "postgres", migrations, migrate.Down)
	if err != nil {
		panic(err)
	}
	_, err = migrate.Exec(db.DB, "postgres", migrations, migrate.Up)
	if err != nil {
		panic(err)
	}
}

func fillDatabase() {
	tx, err := db.Begin(context.Background())
	if err != nil {
		panic(err)
	}
	insertQuery := `INSERT INTO market_user(first_name, last_name, national_code, password)  VALUES ($1, $2, $3, $4)`
	nationalCode := 1234567890
	ct, err := tx.Exec(context.Background(), insertQuery, faker.FirstName(), faker.LastName(), strconv.Itoa(nationalCode))
	if err != nil {
		panic(err)
	}
	if ct.RowsAffected() != 1 {
		panic("affected rows is not 1 and is: " + strconv.Itoa(int(ct.RowsAffected())))
	}
}

func init() {
	var err error

	logger.Infoln(os.Getwd())

	connString := "host=localhost port=5432 user=postgres password=postgres dbname=stock_db_test sslmode=disable"
	db, err = pgxpool.Connect(context.Background(), connString)
	if err != nil {
		panic(err)
	}
	doMigrations(connString)
	//fillDatabase()

	api = server.New(logger, db)
}

func TestMarketUser(t *testing.T) {
	var jwtToken string
	t.Run("create user", func(t *testing.T) {
		tests := []struct {
			name string
			body string
			want struct {
				status int
				body   string
			}
		}{
			{
				name: "binding failed with empty body",
				body: `{}`,
				want: struct {
					status int
					body   string
				}{status: http.StatusBadRequest, body: `error.*validation`},
			},
			{
				name: "new user",
				body: `{
  "FirstName": "John",
  "LastName": "Smith",
  "NationalCode": "1234567890",
  "Password": "P@ssword"
}`,
				want: struct {
					status int
					body   string
				}{status: http.StatusCreated, body: ``},
			},
			{
				name: "new user conflict",
				body: `{
  "FirstName": "John",
  "LastName": "Smith",
  "NationalCode": "1234567890",
  "Password": "P@ssword"
}`,
				want: struct {
					status int
					body   string
				}{status: http.StatusConflict, body: `error.*exists`},
			},
			{
				name: "new user with different name conflict",
				body: `{
  "FirstName": "Jack",
  "LastName": "Smith",
  "NationalCode": "1234567890",
  "Password": "P@ssword1234"
}`,
				want: struct {
					status int
					body   string
				}{status: http.StatusConflict, body: `error.*exists`},
			},
			{
				name: "create more users 1",
				body: `{
  "FirstName": "Jack",
  "LastName": "Bane",
  "NationalCode": "1234567891",
  "Password": "P@ssword1"
}`,
				want: struct {
					status int
					body   string
				}{status: http.StatusCreated, body: ``},
			},
			{
				name: "create more users 2",
				body: `{
  "FirstName": "Jane",
  "LastName": "Goods",
  "NationalCode": "1234567892",
  "Password": "P@ssword2"
}`,
				want: struct {
					status int
					body   string
				}{status: http.StatusCreated, body: ``},
			},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				w := httptest.NewRecorder()
				req, _ := http.NewRequest(http.MethodPost, "/users/", strings.NewReader(tt.body))
				api.ServeHTTP(w, req)

				assert.Equal(t, tt.want.status, w.Code)
				assert.Regexp(t, tt.want.body, w.Body.String())
			})
		}
	})
	t.Run("authenticate", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodPost, "/login", strings.NewReader(`{"NationalCode": "1234567890","Password": "P@ssword"}`))
		api.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Regexp(t, `token.+[A-Za-z0-9-_=]+\.[A-Za-z0-9-_=]+\.?[A-Za-z0-9-_.+/=]*`, w.Body.String())
		authResponse := struct {
			Token string
		}{}
		err := json.NewDecoder(w.Body).Decode(&authResponse)
		if err != nil {
			t.Errorf("decoding failed: %v", err)
		}
		jwtToken = authResponse.Token
		logger.Infoln(jwtToken)
	})
	t.Run("get user", func(t *testing.T) {
		tests := []struct {
			name  string
			token string
			id    int
			want  struct {
				status int
				body   string
			}
		}{
			{
				name:  "successful user retrieval",
				token: jwtToken,
				id:    1,
				want: struct {
					status int
					body   string
				}{
					status: http.StatusOK,
					body:   `NationalCode.*1234567890`,
				},
			},
			{
				name:  "unauthorized access with token",
				token: jwtToken,
				id:    2,
				want: struct {
					status int
					body   string
				}{
					status: http.StatusForbidden,
					body:   `don't have permission`,
				},
			},
			{
				name:  "unauthorized access without token",
				token: "",
				id:    2,
				want: struct {
					status int
					body   string
				}{
					status: http.StatusUnauthorized,
					body:   `auth header is empty`,
				},
			},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				w := httptest.NewRecorder()
				req, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("/users/%d", tt.id), nil)
				if tt.token != "" {
					req.Header.Add("Authorization", "Bearer "+tt.token)
				}
				api.ServeHTTP(w, req)

				assert.Equal(t, tt.want.status, w.Code)
				assert.Regexp(t, tt.want.body, w.Body.String())
			})
		}
	})
}
