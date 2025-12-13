package tunnel_system

import (
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"strconv"
	"time"
)

type TunnelSystem struct {
	mainEntrance *Tunnel
	sideEntrance *Tunnel
}

// VisitorInput represents the JSON structure for incoming visitor events
type VisitorInput struct {
	Topic   string `json:"topic"`
	Payload string `json:"payload"`
}

func NewTunnelSystem(config Config, generators []InputGenerator) {
	// If empty config passed, use defaults
	if config.HTTPPort == "" && config.WebSocketPort == "" && !config.EnableHTTP && !config.EnableWebSocket {
		config = DefaultConfig()
	}

	actionLogger := NewActionLogger()
	mainEntrance := NewNormalTunnel(actionLogger)
	sideEntrance := NewDebugTunnel(actionLogger)
	tunnelSystem := &TunnelSystem{
		mainEntrance: mainEntrance,
		sideEntrance: sideEntrance,
	}

	if config.EnableHTTP {
		if config.HTTPPort == "" {
			config.HTTPPort = ":8081"
		}
		httpGen := createHTTPGenerator(config.HTTPPort)
		generators = append([]InputGenerator{httpGen}, generators...)
	}

	if config.EnableWebSocket {
		if config.WebSocketPort == "" {
			config.WebSocketPort = ":8082"
		}
		wsGen := createWebSocketGenerator(config.WebSocketPort)
		generators = append([]InputGenerator{wsGen}, generators...)
	}

	engineTickGenerator := NewCustomInputGenerator(
		func() VisitorInput {
			return VisitorInput{
				Topic:   string(TICK),
				Payload: strconv.FormatInt(time.Now().UTC().UnixNano(), 10),
			}
		},
		1*time.Second,
	)

	generators = append(generators, engineTickGenerator)

	startInputGenerators(tunnelSystem, generators)

	srv := newTunnelServer(tunnelSystem)
	srv.start(":8080")

	tunnelSystem.openUp()
}

func (t *TunnelSystem) openUp() {
	for {
		v, err := t.mainEntrance.NextVisitor()
		if err != nil {
			fmt.Println(err)
			return
		}
		if v == nil {
			continue
		}
		v, err = t.mainEntrance.Exit(v)
		if err != nil {
			fmt.Println(err)
			return
		}
	}
}
