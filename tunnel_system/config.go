package tunnel_system

// Config for TunnelSystem with optional built-in generators
type Config struct {
	EnableHTTP      bool
	HTTPPort        string
	EnableWebSocket bool
	WebSocketPort   string
}

func DefaultConfig() Config {
	return Config{
		EnableHTTP:      true,
		HTTPPort:        ":8081",
		EnableWebSocket: true,
		WebSocketPort:   ":8082",
	}
}
