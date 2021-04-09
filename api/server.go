package api

import (
	db "github.com/Jiaget/simplebank/db/sqlc"
	"github.com/gin-gonic/gin"
)

// Server serves HTTP requests for the bank service
type Server struct {
	// we make this field pravite, so we can only use the Server.Start() function
	// to call this field
	store  db.Store
	router *gin.Engine
}

// NewServer creates a new HTTP server and setup a router
func NewServer(store db.Store) *Server {
	server := &Server{store: store}
	router := gin.Default()

	router.POST("/accounts", server.createAccount)
	router.GET("/accounts/:id", server.getAccount)
	router.GET("/accounts", server.listAccount)

	server.router = router
	return server
}

// Start runs the HTTP server on a specific address.
func (server *Server) Start(address string) error {
	return server.router.Run(address)
}

// gin.H is a map interface implemented by gin
func errorResponse(err error) gin.H {
	return gin.H{"error": err.Error()}
}
