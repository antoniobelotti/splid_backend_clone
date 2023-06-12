package http

import (
	"github.com/antoniobelotti/splid_backend_clone/internal/group"
	"github.com/antoniobelotti/splid_backend_clone/internal/person"
	"github.com/gin-gonic/gin"
)

type RESTServer struct {
	personService person.Service
	engine        *gin.Engine
}

func NewRESTServer(ps person.Service, gs group.Service) RESTServer {
	router := gin.New()

	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	v1 := router.Group("/api/v1")

	groupHandlers := NewGroupHandlers(gs)
	groupEndpoints := v1.Group("/group")
	{
		groupEndpoints.POST("/", groupHandlers.handleCreateGroup)
	}

	return RESTServer{
		personService: ps,
		engine:        router,
	}
}

func (s *RESTServer) Run(port string) error {
	return s.engine.Run(":" + port)
}
