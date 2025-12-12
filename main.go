package main

import (
	"fmt"
	"fund78/tunnel"
	"fund78/tunnel_system"
	"strconv"
	"time"
)

func main() {
	generators := []tunnel_system.InputGenerator{
		tunnel_system.NewInputGenerator(tunnel.TICK, strconv.FormatInt(time.Now().UTC().UnixNano(), 10), 1*time.Second),
		tunnel_system.NewInputGenerator(tunnel.LOGON, "bob", 5*time.Second),
		tunnel_system.NewCustomInputGenerator(tunnel.TICK, func() string { return fmt.Sprintf("ct-%d", time.Now().Unix()) }, 3*time.Second),
	}

	tunnelSystem := tunnel_system.NewTunnelSystem(tunnel_system.Config{}, generators)
	tunnelSystem.OpenUp()
}
