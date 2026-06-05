package main

import (
	"log"

	"github.com/struckchure/rt-comms/internal"
)

func main() {
	cfg, err := internal.LoadConfig()
	if err != nil {
		log.Fatal(err)
	}
	srv, err := newServer(cfg)
	if err != nil {
		log.Fatal(err)
	}
	if err := srv.run(); err != nil {
		log.Fatal(err)
	}
}
