package main

import (
	"fmt"
	"fund78/tunnel"
	"fund78/tunnel_system"
	_ "github.com/mattn/go-sqlite3"
	"log"
	"time"
)

func main() {
	generators := []tunnel_system.InputGenerator{
		tunnel_system.NewTickGenerator(1 * time.Second),
		tunnel_system.NewStaticGenerator(tunnel.LOGON, "bob", 5*time.Second),
	}

	tunnelSystem := tunnel_system.NewTunnelSystem(generators)

	log.Println("Starting automatic input generators")

	for {
		v, err := tunnelSystem.MainEntrance().NextVisitor()
		if err != nil {
			fmt.Println(err)
			return
		}
		v, err = tunnelSystem.MainEntrance().Exit(v)
		if err != nil {
			fmt.Println(err)
			return
		}
	}
}
