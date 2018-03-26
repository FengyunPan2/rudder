package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/easystack/rudder/src/config"
	"github.com/easystack/rudder/src/router"
)

func main() {
	conf := config.GetConfig()
	log.Printf("http server %s starting...\n", fmt.Sprintf("%s:%s", conf.Address, conf.Port))

	server := &http.Server{
		Addr:           fmt.Sprintf("%s:%s", conf.Address, conf.Port),
		Handler:        router.CreateHTTPRouter(),
		MaxHeaderBytes: 1 << 20,
	}

	err := server.ListenAndServe()
	log.Fatalf("create listen server error: %v", err)
}
