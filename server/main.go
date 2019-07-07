// Copyright © 2019 Adir Tzuberi <adir85@gmail.com>
package main

import (
	"github.com/adirt/rfsb/server/server"
	"log"
	"os/user"
)

func main() {
	currentUser, err := user.Current()
	if err != nil {
		log.Fatalf("failed to get home directory: %v", err)
	}
	rfsbServer, err := server.NewServer(currentUser.HomeDir)
	if err != nil {
		log.Fatalf("failed to initialize server: %v", err)
	}
	if err = rfsbServer.Serve(); err != nil {
		log.Fatalf(err.Error())
	}
}
