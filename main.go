package main

import (
	"fund78/tunnel_system"
)

func main() {
	tunnel_system.NewTunnelSystem(tunnel_system.Config{}, []tunnel_system.InputGenerator{})
}
