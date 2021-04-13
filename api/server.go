package api

import (
	"fmt"

	db "github.com/Jiaget/simplebank/db/sqlc"
	"github.com/Jiaget/simplebank/token"
	"github.com/Jiaget/simplebank/util"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
)

// Server serves HTTP requests for the bank service
type Server struct {
	// we make this field pravite, so we can only use the Server.Start() function
	// to call this field
	store      db.Store
	router     *gin.Engine
	tokenMaker token.Maker
	config     util.Config
}

// NewServer creates a new HTTP server and setup a router
func NewServer(config util.Config, store db.Store) (*Server, error) {
	tokenMaker, err := token.NewPasetoMaker(config.TokenSymmetricKey)
	if err != nil {
		return nil, fmt.Errorf("cannot make a token: %v", err)
	}
	server := &Server{
		config:     config,
		store:      store,
		tokenMaker: tokenMaker,
	}
	router := gin.Default()

	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		v.RegisterValidation("currency", validateCurrency)
	}

	router.POST("/accounts", server.createAccount)
	router.GET("/accounts/:id", server.getAccount)
	router.GET("/accounts", server.listAccount)

	router.POST("/transfer", server.createTransfer)
	router.POST("/users", server.createUser)

	server.router = router
	return server, nil
}

// Start runs the HTTP server on a specific address.
func (server *Server) Start(address string) error {
	return server.router.Run(address)
}

// gin.H is a map interface implemented by gin
func errorResponse(err error) gin.H {
	return gin.H{"error": err.Error()}
}
