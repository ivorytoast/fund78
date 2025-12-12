# Fund78 Svelte Client

Minimal Svelte app to interact with Fund78 HTTP and WebSocket endpoints.

## Setup

```bash
npm install
npm run dev
```

Then open http://localhost:5173

## Features

- **HTTP Section**: Send POST requests to http://localhost:8081/visitor
- **WebSocket Section**: Connect to ws://localhost:8082/ws and send messages

## Usage

1. Start your Go server first: `go run main.go`
2. Start this client: `npm run dev`
3. Use the forms to send HTTP requests or WebSocket messages
