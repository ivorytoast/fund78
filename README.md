# fund78

A minimal Go queue with two modes:
- Engine mode: enqueue events programmatically and write logs under a date-based simulations folder
- Replay mode: read events from simulation logs and replay them

## Quick start

Prerequisites: Go 1.21+

### Run Engine mode (generate a simulation)
```
go run main.go engine
```
Creates logs under:
```
simulations/YYYY/MM/DD/
  input_HHMMSS.log
  output_HHMMSS.log
  input.log        # latest (real file)
  output.log       # latest (real file)
```

### Run Replay mode (replay the latest simulation)
```
go run main.go
```
Finds the latest `simulations/YYYY/MM/DD/input.log` and replays it. Per-run debug logs are written under `simulations/YYYY/MM/DD/debug/` with a timestamp matching the source input.

### Compare debug vs timestamped output
```
go run main.go compare
```
For each `simulations/YYYY/MM/DD/debug/output_debug_HHMMSS.log`, compares against `simulations/YYYY/MM/DD/output_HHMMSS.log` and prints:
- identical | different
- for differences, the first differing line number and both line values

## Event schema
Every event must be a JSON envelope:
```
{
  "topic":   string,          // non-empty
  "payload": string (json)    // must parse as valid JSON
}
```
Examples:
- Valid: `{ "topic":"auth", "payload":"{\"user\":\"alice\"}" }`
- Invalid (payload not JSON): `{ "topic":"auth", "payload":"not json" }`
- Invalid (missing topic): `{ "payload":"{\"k\":\"v\"}" }`

## Folder structure
```
simulations/
  YYYY/MM/DD/
    input_HHMMSS.log       # Engine per-run input
    output_HHMMSS.log      # Engine per-run output
    input.log              # latest input (real file)
    output.log             # latest output (real file)
    debug/
      input_debug_HHMMSS.log    # Replay per-run debug (input mirror)
      output_debug_HHMMSS.log   # Replay per-run debug (output mirror)
```

## Environment variables
- `SIMULATIONS_DIR`: override base folder for simulations (default: `simulations`)
  (UI is optional; if used: `PORT` controls server port, default `8080`)

## Development

### Run tests
```
go test ./internal/queue -v
```
Tests run with a temporary `SIMULATIONS_DIR` and remove all files/folders after completion.

### Commands recap
- Engine: `go run main.go engine`
- Replay: `go run main.go`
- Compare: `go run main.go compare`

## Notes
- Engine writes both timestamped logs and latest real files.
- Replay writes per-run debug logs with timestamps matching the source input.

```go
generators := []tunnel_system.InputGenerator{
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
```
