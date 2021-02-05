package server

import (
	"context"
	"log"
	"os"

	"github.com/alexedwards/argon2id"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v4"
)

type Server struct {
	db         *pgx.Conn
	router     *gin.Engine
	hashParams *argon2id.Params
}

func NewServer() *Server {
	s := &Server{}
	datasource := os.ExpandEnv("host=127.0.0.1 port=${POSTGRES_PORT} user=${POSTGRES_USER} password=${POSTGRES_PASSWORD} dbname=${DATABASE_NAME} sslmode=disable")

	var err error
	if s.db, err = pgx.Connect(context.Background(), datasource); err != nil {
		log.Fatal(err)
	}

	s.hashParams = &argon2id.Params{
		Memory:      64 * 1024,
		Iterations:  1,
		Parallelism: 2,
		SaltLength:  16,
		KeyLength:   32,
	}

	s.router = s.NewAPI()
	return s
}

func (s *Server) NewAPI() *gin.Engine {
	app := gin.Default()

	api := app.Group("/api")
	api.POST("register", s.registerHandler)

	return app
}

func (s *Server) Run(addr ...string) error {
	return s.router.Run(addr...)
}
