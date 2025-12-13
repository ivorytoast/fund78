package tunnel_system

import (
	"encoding/json"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"time"
)

// InputGenerator is the interface that both generator types implement
type InputGenerator interface {
	Start(*TunnelSystem)
}

// IntervalGenerator generates events at fixed time intervals
type IntervalGenerator struct {
	InputFunc func() VisitorInput
	Interval  time.Duration
}

// ConnectionGenerator generates events from external connections
type ConnectionGenerator struct {
	StartFunc func(*Tunnel)
}

func (g *IntervalGenerator) Start(ts *TunnelSystem) {
	go func() {
		for {
			input := g.InputFunc()
			v := NewInputAction(ActionName(input.Topic), input.Payload)
			ts.mainEntrance.Enter(v)
			time.Sleep(g.Interval)
		}
	}()
	log.Printf("Started interval generator (interval: %v)", g.Interval)
}

func (g *ConnectionGenerator) Start(ts *TunnelSystem) {
	go func() {
		log.Printf("Started connection generator")
		g.StartFunc(ts.mainEntrance)
	}()
}

func NewCustomInputGenerator(inputFunc func() VisitorInput, interval time.Duration) *IntervalGenerator {
	return &IntervalGenerator{
		InputFunc: inputFunc,
		Interval:  interval,
	}
}

func NewConnectionInputGenerator(startFunc func(*Tunnel)) *ConnectionGenerator {
	return &ConnectionGenerator{
		StartFunc: startFunc,
	}
}

func startInputGenerators(ts *TunnelSystem, generators []InputGenerator) {
	for _, gen := range generators {
		gen.Start(ts)
	}
}

// createHTTPGenerator creates a built-in HTTP server generator
func createHTTPGenerator(port string) *ConnectionGenerator {
	return NewConnectionInputGenerator(func(t *Tunnel) {
		mux := http.NewServeMux()

		mux.HandleFunc("/visitor", func(w http.ResponseWriter, r *http.Request) {
			// Set CORS headers
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

			// Handle preflight OPTIONS request
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusOK)
				return
			}

			if r.Method != http.MethodPost {
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
				return
			}

			var input VisitorInput
			if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
				http.Error(w, "Invalid JSON", http.StatusBadRequest)
				return
			}

			if input.Topic == "" {
				http.Error(w, "Topic is required", http.StatusBadRequest)
				return
			}

			v := NewInputAction(ActionName(input.Topic), input.Payload)
			t.Enter(v)

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]string{"status": "ok"})

			log.Printf("HTTP: Received %s with payload: %s", input.Topic, input.Payload)
		})

		log.Printf("HTTP server listening on %s/visitor", port)
		if err := http.ListenAndServe(port, mux); err != nil {
			log.Printf("HTTP server error: %v", err)
		}
	})
}

// createWebSocketGenerator creates a built-in WebSocket server generator
func createWebSocketGenerator(port string) *ConnectionGenerator {
	return NewConnectionInputGenerator(func(t *Tunnel) {
		upgrader := websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true // Allow all origins
			},
		}

		mux := http.NewServeMux()

		mux.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
			conn, err := upgrader.Upgrade(w, r, nil)
			if err != nil {
				log.Printf("WebSocket upgrade failed: %v", err)
				return
			}
			defer conn.Close()

			log.Printf("WebSocket client connected")

			for {
				_, message, err := conn.ReadMessage()
				if err != nil {
					log.Printf("WebSocket read error: %v", err)
					break
				}

				var input VisitorInput
				if err := json.Unmarshal(message, &input); err != nil {
					log.Printf("WebSocket invalid JSON: %v", err)
					continue
				}

				if input.Topic == "" {
					log.Printf("WebSocket: Topic is required")
					continue
				}

				v := NewInputAction(ActionName(input.Topic), input.Payload)
				t.Enter(v)

				log.Printf("WebSocket: Received %s with payload: %s", input.Topic, input.Payload)
			}
		})

		log.Printf("WebSocket server listening on %s/ws", port)
		if err := http.ListenAndServe(port, mux); err != nil {
			log.Printf("WebSocket server error: %v", err)
		}
	})
}
