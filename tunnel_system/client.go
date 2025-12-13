package tunnel_system

import (
	"fmt"
	"log"
	"net/http"
)

type server struct {
	tunnelSystem *TunnelSystem
}

func newTunnelServer(tunnelSystem *TunnelSystem) *server {
	return &server{
		tunnelSystem: tunnelSystem,
	}
}

func (s *server) start(port string) {
	http.HandleFunc("/replay/", s.handleGetReplay)
	http.HandleFunc("/rerun/", s.handleRerunReplay)
	http.HandleFunc("/compare/", s.handleCompareReplay)

	log.Printf("Server starting on http://localhost%s", port)
	log.Printf("Get replay actions: GET http://localhost%s/replay/{id}", port)
	log.Printf("Re-run replay: GET http://localhost%s/rerun/{id}", port)
	log.Printf("Compare replay to debug runs: GET http://localhost%s/compare/{id}", port)

	go func() {
		if err := http.ListenAndServe(port, nil); err != nil {
			log.Fatal(err)
		}
	}()
}

func (s *server) handleGetReplay(w http.ResponseWriter, r *http.Request) {
	//// Extract replay ID from URL path
	//var replayID int64
	//path := r.URL.Path
	//_, err := fmt.Sscanf(path, "/replay/%d", &replayID)
	//if err != nil {
	//	http.Error(w, "Invalid replay ID. Use /replay/{id}", http.StatusBadRequest)
	//	return
	//}
	//
	//// Get messages for this replay (already ordered from first to most recent)
	//messages, err := s.tunnelSystem.mainEntrance().GetMessagesByReplayID(replayID)
	//if err != nil {
	//	http.Error(w, fmt.Sprintf("Error fetching messages: %v", err), http.StatusInternalServerError)
	//	return
	//}
	//
	//// Print them out in order
	//w.Header().Set("Content-Type", "text/plain")
	//fmt.Fprintf(w, "Replay ID: %d\n", replayID)
	//fmt.Fprintf(w, "Total Actions: %d\n", len(messages))
	//fmt.Fprintf(w, "%s\n\n", "===========================================")
	//
	//for i, msg := range messages {
	//	fmt.Fprintf(w, "[%d] Message ID: %s\n", i+1, msg.MessageID)
	//	fmt.Fprintf(w, "    ActionName: %s\n", msg.Topic)
	//	fmt.Fprintf(w, "    Type: %s | ActionDirection: %s | Action Type: %s\n", msg.MessageType, msg.Direction, msg.ActionType)
	//	fmt.Fprintf(w, "    Caused By: %s\n", msg.CausedBy)
	//	fmt.Fprintf(w, "    Payload: %s\n", msg.Payload)
	//	fmt.Fprintf(w, "    Created At: %d\n", msg.CreatedAt)
	//	fmt.Fprintf(w, "\n")
	//}
}

func (s *server) handleRerunReplay(w http.ResponseWriter, r *http.Request) {
	//// Extract replay ID from URL path
	//var replayID int64
	//path := r.URL.Path
	//_, err := fmt.Sscanf(path, "/rerun/%d", &replayID)
	//if err != nil {
	//	http.Error(w, "Invalid replay ID. Use /rerun/{id}", http.StatusBadRequest)
	//	return
	//}
	//
	//// Get optional name from query parameter
	//debugName := r.URL.Query().Get("name")
	//if debugName == "" {
	//	debugName = fmt.Sprintf("Debug of replay %d", replayID)
	//}
	//
	//// Create a new replay entry as a child of the original
	//debugReplayID, err := s.explorer.InsertReplay(debugName, "", 1, &replayID)
	//if err != nil {
	//	http.Error(w, fmt.Sprintf("Error creating debug replay: %v", err), http.StatusInternalServerError)
	//	return
	//}
	//
	//// Get messages for this replay (ordered from first to most recent)
	//messages, err := s.explorer.GetMessagesByReplayID(replayID)
	//if err != nil {
	//	http.Error(w, fmt.Sprintf("Error fetching messages: %v", err), http.StatusInternalServerError)
	//	return
	//}
	//
	//// Re-enqueue all messages in order with the NEW debug replay ID
	//for _, msg := range messages {
	//	visitor := tunnel.NewVisitorFromActionRow(
	//		msg.MessageID,
	//		msg.Topic,
	//		msg.CausedBy,
	//		msg.MessageType,
	//		msg.Direction,
	//		msg.Payload,
	//		debugReplayID,
	//	)
	//	s.path.Enter(visitor)
	//}
	//
	//// Send response
	//w.Header().Set("Content-Type", "text/plain")
	//fmt.Fprintf(w, "Created debug replay %d (parent: %d) named '%s'\n", debugReplayID, replayID, debugName)
	//fmt.Fprintf(w, "Successfully re-queued %d messages\n", len(messages))
	//fmt.Fprintf(w, "Messages will be processed by the main loop\n")
}

// Alternative handler if you prefer query parameter instead of path parameter
// Usage: /replay?id=123
func (s *server) handleGetReplayQuery(w http.ResponseWriter, r *http.Request) {
	//idStr := r.URL.Query().Get("id")
	//if idStr == "" {
	//	http.Error(w, "Missing 'id' parameter. Use /replay?id={id}", http.StatusBadRequest)
	//	return
	//}
	//
	//replayID, err := strconv.ParseInt(idStr, 10, 64)
	//if err != nil {
	//	http.Error(w, "Invalid replay ID", http.StatusBadRequest)
	//	return
	//}
	//
	//messages, err := s.explorer.GetMessagesByReplayID(replayID)
	//if err != nil {
	//	http.Error(w, fmt.Sprintf("Error fetching messages: %v", err), http.StatusInternalServerError)
	//	return
	//}
	//
	//w.Header().Set("Content-Type", "application/json")
	//if err := json.NewEncoder(w).Encode(messages); err != nil {
	//	http.Error(w, fmt.Sprintf("Error encoding response: %v", err), http.StatusInternalServerError)
	//	return
	//}
}

type ComparisonResult struct {
	OriginalReplayID int64                `json:"original_replay_id"`
	OriginalName     string               `json:"original_name"`
	ActionCount      int                  `json:"action_count"`
	DebugRuns        []DebugRunComparison `json:"debug_runs"`
}

type DebugRunComparison struct {
	ReplayID    int64    `json:"replay_id"`
	Name        string   `json:"name"`
	ActionCount int      `json:"action_count"`
	Identical   bool     `json:"identical"`
	Differences []string `json:"differences,omitempty"`
}

func (s *server) handleCompareReplay(w http.ResponseWriter, r *http.Request) {
	//// Extract replay ID from URL path
	//var replayID int64
	//path := r.URL.Path
	//_, err := fmt.Sscanf(path, "/compare/%d", &replayID)
	//if err != nil {
	//	http.Error(w, "Invalid replay ID. Use /compare/{id}", http.StatusBadRequest)
	//	return
	//}
	//
	//// Get the original replay info
	//allReplays, err := s.explorer.GetAllReplays()
	//if err != nil {
	//	http.Error(w, fmt.Sprintf("Error fetching replays: %v", err), http.StatusInternalServerError)
	//	return
	//}
	//
	//var originalReplay *action_logger.Replay
	//for _, r := range allReplays {
	//	if r.ID == replayID {
	//		originalReplay = &r
	//		break
	//	}
	//}
	//
	//if originalReplay == nil {
	//	http.Error(w, fmt.Sprintf("Replay %d not found", replayID), http.StatusNotFound)
	//	return
	//}
	//
	//// Get original actions
	//originalActions, err := s.explorer.GetMessagesByReplayID(replayID)
	//if err != nil {
	//	http.Error(w, fmt.Sprintf("Error fetching original actions: %v", err), http.StatusInternalServerError)
	//	return
	//}
	//
	//// Get all child replays (debug runs)
	//childReplays, err := s.explorer.GetChildReplays(replayID)
	//if err != nil {
	//	http.Error(w, fmt.Sprintf("Error fetching child replays: %v", err), http.StatusInternalServerError)
	//	return
	//}
	//
	//// Compare each debug run to the original
	//debugRuns := make([]DebugRunComparison, 0)
	//for _, child := range childReplays {
	//	debugActions, err := s.explorer.GetMessagesByReplayID(child.ID)
	//	if err != nil {
	//		http.Error(w, fmt.Sprintf("Error fetching debug actions: %v", err), http.StatusInternalServerError)
	//		return
	//	}
	//
	//	comparison := compareActions(originalActions, debugActions)
	//	debugRuns = append(debugRuns, DebugRunComparison{
	//		ReplayID:    child.ID,
	//		Name:        child.Name,
	//		ActionCount: len(debugActions),
	//		Identical:   comparison.Identical,
	//		Differences: comparison.Differences,
	//	})
	//}
	//
	//result := ComparisonResult{
	//	OriginalReplayID: replayID,
	//	OriginalName:     originalReplay.Name,
	//	ActionCount:      len(originalActions),
	//	DebugRuns:        debugRuns,
	//}
	//
	//w.Header().Set("Content-Type", "application/json")
	//if err := json.NewEncoder(w).Encode(result); err != nil {
	//	http.Error(w, fmt.Sprintf("Error encoding response: %v", err), http.StatusInternalServerError)
	//	return
	//}
}

type actionComparison struct {
	Identical   bool
	Differences []string
}

func compareActions(original, debug []ActionRow) actionComparison {
	differences := make([]string, 0)

	// Check count
	if len(original) != len(debug) {
		differences = append(differences, fmt.Sprintf("Action count mismatch: original has %d, debug has %d", len(original), len(debug)))

		// Still compare the common actions
		minLen := len(original)
		if len(debug) < minLen {
			minLen = len(debug)
		}

		for i := 0; i < minLen; i++ {
			diff := compareMessage(original[i], debug[i], i)
			differences = append(differences, diff...)
		}

		return actionComparison{
			Identical:   false,
			Differences: differences,
		}
	}

	// Compare each action
	for i := 0; i < len(original); i++ {
		diff := compareMessage(original[i], debug[i], i)
		differences = append(differences, diff...)
	}

	return actionComparison{
		Identical:   len(differences) == 0,
		Differences: differences,
	}
}

func compareMessage(orig, dbg ActionRow, index int) []string {
	differences := make([]string, 0)

	if orig.MessageID != dbg.MessageID {
		differences = append(differences, fmt.Sprintf("Index %d: MessageID differs (%s vs %s)", index, orig.MessageID, dbg.MessageID))
	}
	if orig.Topic != dbg.Topic {
		differences = append(differences, fmt.Sprintf("Index %d: ActionName differs (%s vs %s)", index, orig.Topic, dbg.Topic))
	}
	if orig.CausedBy != dbg.CausedBy {
		differences = append(differences, fmt.Sprintf("Index %d: CausedBy differs (%s vs %s)", index, orig.CausedBy, dbg.CausedBy))
	}
	if orig.MessageType != dbg.MessageType {
		differences = append(differences, fmt.Sprintf("Index %d: ActionType differs (%s vs %s)", index, orig.MessageType, dbg.MessageType))
	}
	if orig.Direction != dbg.Direction {
		differences = append(differences, fmt.Sprintf("Index %d: ActionDirection differs (%s vs %s)", index, orig.Direction, dbg.Direction))
	}
	if orig.Payload != dbg.Payload {
		differences = append(differences, fmt.Sprintf("Index %d: Payload differs (%s vs %s)", index, orig.Payload, dbg.Payload))
	}

	return differences
}
