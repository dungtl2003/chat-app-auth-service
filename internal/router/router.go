package router

import (
	"log"
	"log/slog"

	"github.com/gin-gonic/gin"
	sloggin "github.com/samber/slog-gin"
)

type Method string

const (
	GET    Method = "GET"
	POST   Method = "POST"
	PUT    Method = "PUT"
	PATCH  Method = "PATCH"
	DELETE Method = "DELETE"
)

type Handler struct {
	Method Method
	Path   string
	H      gin.HandlerFunc
}

func isMethodValid(m Method) bool {
	switch m {
	case GET, POST, PUT, DELETE, PATCH:
		return true
	default:
		return false
	}
}

func New(logger *slog.Logger, handlers ...Handler) *gin.Engine {
	r := gin.New()
	r.Use(sloggin.New(logger))
	r.Use(gin.Logger())
	r.Use(gin.Recovery())

	for _, h := range handlers {
		if !isMethodValid(h.Method) {
			log.Fatalf("Invalid method: %s", h.Method)
		}

		r.Handle(string(h.Method), h.Path, h.H)
	}

	return r
}
