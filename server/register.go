package server

import (
	"github.com/alexedwards/argon2id"
	"github.com/gin-gonic/gin"
)

func (s *Server) registerHandler(c *gin.Context) {
	data := struct {
		Password string `json:"password"`
	}{}
	c.BindJSON(&data)

	argon2id.CreateHash(data.Password, s.hashParams)
}
