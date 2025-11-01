package main

import (
	"log"

	"github.com/leledumbo/fiskaly-coding-challenge/signing-service-challenge-go/api/server"
)

const (
	ListenAddress = ":8080"
	// TODO: add further configuration parameters here ...
)

func main() {
	s := server.NewServer(ListenAddress)

	if err := s.Run(); err != nil {
		log.Fatal("Could not start server on ", ListenAddress)
	}
}
