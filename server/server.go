package server

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v4"
	"log"
	"os"
)

type Server struct {
	db     *pgx.Conn
	router *gin.Engine
}

func NewServer() *Server {
	s := &Server{}
	datasource := os.ExpandEnv("host=127.0.0.1 port=${POSTGRES_PORT} user=${POSTGRES_USER} password=${POSTGRES_PASSWORD} dbname=${DATABASE_NAME} sslmode=disable")

	var err error
	if s.db, err = pgx.Connect(context.Background(), datasource); err != nil {
		log.Fatal(err)
	}

	s.router = NewAPI()
	return s
}

func NewAPI() *gin.Engine {
	app := gin.Default()

	return app
}
