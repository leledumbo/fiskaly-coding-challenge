package server

import (
	"github.com/leledumbo/fiskaly-coding-challenge/signing-service-challenge-go/api/common"
	_ "github.com/leledumbo/fiskaly-coding-challenge/signing-service-challenge-go/api/routes" // we only need to call init() functions in files inside the package
	"net/http"
)

// Server manages HTTP requests and dispatches them to the appropriate services.
type Server struct {
	listenAddress string
}

// NewServer is a factory to instantiate a new Server.
func NewServer(listenAddress string) *Server {
	return &Server{
		listenAddress: listenAddress,
		// TODO: add services / further dependencies here ...
	}
}

// Run starts the Server.
func (s *Server) Run() error {
	return http.ListenAndServe(s.listenAddress, common.Mux())
}
