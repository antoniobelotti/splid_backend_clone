package http

import (
	"github.com/antoniobelotti/splid_backend_clone/internal/expense"
	"github.com/antoniobelotti/splid_backend_clone/internal/group"
	"github.com/antoniobelotti/splid_backend_clone/internal/http/authentication"
	"github.com/antoniobelotti/splid_backend_clone/internal/person"
	"github.com/antoniobelotti/splid_backend_clone/internal/transfer"
	"github.com/gin-gonic/gin"
)

type RESTServer struct {
	*gin.Engine
}

func NewRESTServer(ps person.Service, gs group.Service, es expense.Service, ts transfer.Service) RESTServer {
	router := gin.New()

	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	v1 := router.Group("/api/v1")

	groupHandlers := NewGroupHandlers(gs)
	groupEndpoints := v1.Group("/group")
	{
		groupEndpoints.POST("", authentication.AuthenticateMiddleware(), groupHandlers.handleCreateGroup)
		groupEndpoints.POST("/:groupId/join", authentication.AuthenticateMiddleware(), groupHandlers.handleJoinGroup)
		groupEndpoints.GET("/:groupId/balance", authentication.AuthenticateMiddleware(), groupHandlers.handleGetBalance)
		groupEndpoints.GET("/:groupId/operations-to-even-balance", authentication.AuthenticateMiddleware(), groupHandlers.handleGetOpsEvenBalance)
	}

	personHandlers := NewPersonHandlers(ps)
	personEndpoints := v1.Group("/person")
	{
		personEndpoints.POST("/signup", personHandlers.handleCreatePerson)
		personEndpoints.POST("/login", personHandlers.handleLogin)
		personEndpoints.GET("", authentication.AuthenticateMiddleware(), personHandlers.handleGetPerson)
	}

	expenseHandlers := NewExpenseHandlers(es)
	expenseEndpoints := v1.Group("/expense")
	{
		expenseEndpoints.POST("", authentication.AuthenticateMiddleware(), expenseHandlers.handleCreateExpense)
	}

	transferHandlers := NewTransferHandlers(ts)
	transferEndpoints := v1.Group("/transfer")
	{
		transferEndpoints.POST("", authentication.AuthenticateMiddleware(), transferHandlers.handleCreateTransfer)
	}

	return RESTServer{
		router,
	}
}
