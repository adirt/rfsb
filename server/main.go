// Copyright © 2019 Adir Tzuberi <adir85@gmail.com>
package main

import (
	"github.com/adirt/rfsb/server/server"
	"log"
)

func main() {
	rfsbServer := server.RfsbServer{}
	if err := rfsbServer.Serve(); err != nil {
		log.Fatalf(err.Error())
	}
}
