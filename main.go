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
		tunnel_system.NewInputGenerator(
			tunnel_system.VisitorInput{
				Topic:   string(tunnel.TICK),
				Payload: strconv.FormatInt(time.Now().UTC().UnixNano(), 10),
			},
			1*time.Second,
		),
		tunnel_system.NewInputGenerator(
			tunnel_system.VisitorInput{
				Topic:   string(tunnel.LOGON),
				Payload: "bob",
			},
			5*time.Second,
		),
		tunnel_system.NewCustomInputGenerator(
			func() tunnel_system.VisitorInput {
				return tunnel_system.VisitorInput{
					Topic:   string(tunnel.TICK),
					Payload: fmt.Sprintf("ct-%d", time.Now().Unix()),
				}
			},
			3*time.Second,
		),
		// Example ConnectionGenerator - simulates an external system
		tunnel_system.NewConnectionInputGenerator(func(t *tunnel.Tunnel) {
			// Simulate external connection logic
			ticker := time.NewTicker(7 * time.Second)
			counter := 0
			for range ticker.C {
				counter++
				input := tunnel_system.VisitorInput{
					Topic:   string(tunnel.LOGON),
					Payload: fmt.Sprintf("user-%d", counter),
				}
				v := tunnel.NewInputAction(tunnel.ActionName(input.Topic), input.Payload)
				t.Enter(v)
			}
		}),
	}

	tunnelSystem := tunnel_system.NewTunnelSystem(tunnel_system.Config{}, generators)
	tunnelSystem.OpenUp()
}
