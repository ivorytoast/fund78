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
Finds the latest `simulations/YYYY/MM/DD/input.log` and replays it.

### Run the UI and send events (Engine mode)
```
go run cmd/ui/main.go
```
Open http://localhost:8080 and submit:
- topic (string)
- payload (JSON string), e.g. {"k":"v"}

Or via curl:
```
curl -X POST http://localhost:8080/send \
  -d 'topic=auth' \
  --data-urlencode 'payload={"user":"alice","action":"login"}'
```

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
      input_debug_HHMMSS.log    # Replay per-run debug
      output_debug_HHMMSS.log
```

## Environment variables
- `SIMULATIONS_DIR`: override base folder for simulations (default: `simulations`)
- `PORT`: UI server port (default: `8080`)

## Development

### Run tests
```
go test ./internal/queue -v
```
Tests run with a temporary `SIMULATIONS_DIR` and remove all files/folders after completion.

### Commands recap
- Engine: `go run main.go engine`
- Replay: `go run main.go`
- UI: `go run cmd/ui/main.go` then open http://localhost:8080

## Notes
- Engine writes both timestamped logs and latest real files.
- Replay writes per-run debug logs with timestamps matching the source input.
