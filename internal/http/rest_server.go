package http

import (
	"github.com/antoniobelotti/splid_backend_clone/internal/group"
	"github.com/antoniobelotti/splid_backend_clone/internal/person"
	"github.com/gin-gonic/gin"
)

type RESTServer struct {
	*gin.Engine
}

func NewRESTServer(ps person.Service, gs group.Service) RESTServer {
	router := gin.New()

	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	v1 := router.Group("/api/v1")

	groupHandlers := NewGroupHandlers(gs)
	groupEndpoints := v1.Group("/group")
	{
		groupEndpoints.POST("", groupHandlers.handleCreateGroup)
	}

	personHandlers := NewPersonHandlers(ps)
	personEndpoints := v1.Group("/person")
	{
		personEndpoints.POST("", personHandlers.handleCreatePerson)
		personEndpoints.GET("/:personId", personHandlers.handleGetPerson)
		personEndpoints.POST("/login", personHandlers.handleLogin)
	}

	return RESTServer{
		router,
	}
}
