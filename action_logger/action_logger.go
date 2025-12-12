package action_logger

import (
	"database/sql"
	"log"
)

type ActionLogger struct {
	db *sql.DB
}

type Replay struct {
	ID             int64  `json:"id"`
	Name           string `json:"name"`
	FileID         string `json:"file_id"`
	Version        int    `json:"version"`
	ParentReplayID *int64 `json:"parent_replay_id,omitempty"`
	CreatedAt      int64  `json:"created_at"`
}

type ActionRow struct {
	ID          int64  `json:"id"`
	ReplayID    int64  `json:"replay_id"`
	MessageID   string `json:"message_id"`
	Topic       string `json:"topic"`
	CausedBy    string `json:"caused_by"`
	MessageType string `json:"message_type"`
	Direction   string `json:"direction"`
	Payload     string `json:"payload"`
	ActionType  string `json:"action_type"`
	CreatedAt   int64  `json:"created_at"`
}

func NewActionLogger() *ActionLogger {
	db, err := sql.Open("sqlite3", "./fund78db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	sqlText := `
CREATE TABLE IF NOT EXISTS replay_input (
    id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    file_id TEXT NOT NULL,
    version INTEGER NOT NULL,
    parent_replay_id INTEGER,
    created_at INTEGER DEFAULT (strftime('%s','now')) NOT NULL,
    FOREIGN KEY (parent_replay_id) REFERENCES replay_input(id)
);
`
	_, err = db.Exec(sqlText)
	if err != nil {
		panic("the create table statement for replay_input failed because: " + err.Error())
	}

	actionSql := `
CREATE TABLE IF NOT EXISTS action (
    id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
    replay_id INTEGER NOT NULL,
    message_id TEXT NOT NULL,
    topic TEXT NOT NULL,
    caused_by TEXT NOT NULL,
    message_type TEXT NOT NULL,
    direction TEXT NOT NULL,
    payload TEXT NOT NULL,
    action_type TEXT NOT NULL,
    created_at INTEGER DEFAULT (strftime('%s','now')) NOT NULL
);
`

	_, err = db.Exec(actionSql)
	if err != nil {
		panic("the create table statement for action failed because: " + err.Error())
	}

	return &ActionLogger{db: db}
}

func (fx *ActionLogger) InsertAction(replayId int64, messageId string, topic string, causedBy string, messageType string, direction string, payload string, actionType string) {
	sqlText := "INSERT INTO action (replay_id, message_id, topic, caused_by, message_type, direction, payload, action_type) VALUES (?, ?, ?, ?, ?, ?, ?, ?);"
	_, err := fx.db.Exec(sqlText, replayId, messageId, topic, causedBy, messageType, direction, payload, actionType)
	if err != nil {
		log.Fatal(err)
	}
}

func (fx *ActionLogger) InsertReplay(name string, fileId string, version int, parentReplayId *int64) (int64, error) {
	sqlText := "INSERT INTO replay_input (name, file_id, version, parent_replay_id) VALUES (?, ?, ?, ?);"
	result, err := fx.db.Exec(sqlText, name, fileId, version, parentReplayId)
	if err != nil {
		log.Fatal(err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		log.Fatal(err)
	}

	return id, nil
}

func (fx *ActionLogger) GetAllReplays() ([]Replay, error) {
	sqlText := "SELECT id, name, file_id, version, parent_replay_id, created_at FROM replay_input ORDER BY created_at DESC;"
	rows, err := fx.db.Query(sqlText)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	replays := make([]Replay, 0)
	for rows.Next() {
		var replay Replay
		err = rows.Scan(&replay.ID, &replay.Name, &replay.FileID, &replay.Version, &replay.ParentReplayID, &replay.CreatedAt)
		if err != nil {
			return nil, err
		}
		replays = append(replays, replay)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return replays, nil
}

func (fx *ActionLogger) GetChildReplays(parentReplayID int64) ([]Replay, error) {
	sqlText := "SELECT id, name, file_id, version, parent_replay_id, created_at FROM replay_input WHERE parent_replay_id = ? ORDER BY created_at ASC;"
	rows, err := fx.db.Query(sqlText, parentReplayID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	replays := make([]Replay, 0)
	for rows.Next() {
		var replay Replay
		err = rows.Scan(&replay.ID, &replay.Name, &replay.FileID, &replay.Version, &replay.ParentReplayID, &replay.CreatedAt)
		if err != nil {
			return nil, err
		}
		replays = append(replays, replay)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return replays, nil
}

func (fx *ActionLogger) PrintAllReplays() {
	replays, err := fx.GetAllReplays()
	if err != nil {
		log.Fatal(err)
	}
	for _, replay := range replays {
		log.Printf("Replay Input: %d, %s %s", replay.ID, replay.Name, replay.FileID)
	}
}

func (fx *ActionLogger) GetRecentMessages(limit int) ([]ActionRow, error) {
	sqlText := "SELECT id, replay_id, message_id, topic, caused_by, message_type, direction, payload, action_type, created_at FROM action ORDER BY id DESC LIMIT ?;"
	rows, err := fx.db.Query(sqlText, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	messages := make([]ActionRow, 0)
	for rows.Next() {
		var msg ActionRow
		err = rows.Scan(&msg.ID, &msg.ReplayID, &msg.MessageID, &msg.Topic, &msg.CausedBy, &msg.MessageType, &msg.Direction, &msg.Payload, &msg.ActionType, &msg.CreatedAt)
		if err != nil {
			return nil, err
		}
		messages = append(messages, msg)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}

	// Reverse the slice to get chronological order
	for i, j := 0, len(messages)-1; i < j; i, j = i+1, j-1 {
		messages[i], messages[j] = messages[j], messages[i]
	}

	return messages, nil
}

func (fx *ActionLogger) GetMessagesByReplayID(replayID int64) ([]ActionRow, error) {
	sqlText := "SELECT id, replay_id, message_id, topic, caused_by, message_type, direction, payload, action_type, created_at FROM action WHERE replay_id = ? ORDER BY id ASC;"
	rows, err := fx.db.Query(sqlText, replayID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	messages := make([]ActionRow, 0)
	for rows.Next() {
		var msg ActionRow
		err = rows.Scan(&msg.ID, &msg.ReplayID, &msg.MessageID, &msg.Topic, &msg.CausedBy, &msg.MessageType, &msg.Direction, &msg.Payload, &msg.ActionType, &msg.CreatedAt)
		if err != nil {
			return nil, err
		}
		messages = append(messages, msg)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return messages, nil
}
