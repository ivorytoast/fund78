package main

import (
	"encoding/json"
	"fund78/internal/queue"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

type envelope struct {
	Topic   string `json:"topic"`
	Payload string `json:"payload"`
}

var page = template.Must(template.New("index").Parse(`<!DOCTYPE html><html><head><meta charset="utf-8"><title>fund78</title><meta name="viewport" content="width=device-width,initial-scale=1"><style>body{font-family:system-ui,-apple-system,Segoe UI,Roboto,Helvetica,Arial,sans-serif;margin:0;padding:24px;background:#0b0b0c;color:#fff}form{max-width:640px;margin:0 auto;display:flex;flex-direction:column;gap:12px}input,textarea,button{font-size:16px;padding:10px 12px;border-radius:8px;border:1px solid #2a2a2e;background:#141416;color:#fff}button{background:#3b82f6;border:none;cursor:pointer}button:hover{background:#2563eb}.wrap{max-width:720px;margin:0 auto}.card{background:#111114;border:1px solid #232329;border-radius:12px;padding:16px}.row{display:flex;gap:12px}.row>*{flex:1}.mono{font-family:ui-monospace,SFMono-Regular,Menlo,Monaco,Consolas,monospace;font-size:14px;white-space:pre-wrap;background:#0f0f11;border-radius:8px;border:1px solid #26262b;padding:12px}</style></head><body><div class="wrap"><h1>fund78</h1><div class="card"><form method="post" action="/send"><div class="row"><input name="topic" placeholder="topic" required></div><textarea name="payload" rows="6" placeholder='payload JSON, e.g. {"k":"v"}' required></textarea><button type="submit">Send</button></form></div>{{if .Message}}<p>{{.Message}}</p>{{end}}{{if .Last}}<h3>Last event</h3><div class="mono">{{.Last}}</div>{{end}}</div></body></html>`))

func main() {
	q := queue.NewEngineQueue()
	defer q.Stop()

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		_ = page.Execute(w, nil)
	})

	http.HandleFunc("/send", func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			http.Error(w, "bad form", http.StatusBadRequest)
			return
		}
		topic := r.FormValue("topic")
		payload := r.FormValue("payload")
		env := envelope{Topic: topic, Payload: payload}
		b, _ := json.Marshal(env)
		if err := q.Enqueue(string(b)); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		_ = page.Execute(w, map[string]any{"Message": "sent", "Last": string(b)})
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	srv := &http.Server{Addr: ":" + port, ReadHeaderTimeout: 5 * time.Second}
	base := os.Getenv("SIMULATIONS_DIR")
	if base == "" {
		base = "simulations"
	}
	_ = os.MkdirAll(filepath.Join(base), 0755)
	log.Printf("ui listening on :%s", port)
	log.Fatal(srv.ListenAndServe())
}
