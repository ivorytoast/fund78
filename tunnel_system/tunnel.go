package tunnel_system

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"fund78/assert"
	"log"
	"math/big"
)

type ActionType string

const (
	INPUT   ActionType = "INPUT"
	REQUEST ActionType = "REQUEST"
	REPLY   ActionType = "REPLY"
)

type ActionDirection string

const (
	IN  ActionDirection = "IN"
	OUT ActionDirection = "OUT"
)

type ActionName string

const (
	TICK  ActionName = "TICK"
	LOGON ActionName = "LOGON"
)

type Tunnel struct {
	queue        chan *Visitor
	replayId     int64
	actionLogger *ActionLogger
}

type Visitor struct {
	MessageId       string          `json:"messageId"`
	ActionName      ActionName      `json:"actionName"`
	CausedBy        string          `json:"causedBy"`
	Payload         string          `json:"payload"`
	ActionType      ActionType      `json:"actionType"`
	ActionDirection ActionDirection `json:"actionDirection"`
	IsDebug         bool            `json:"isDebug"`
	ReplayId        int64           `json:"replayId"`
}

func (t *Tunnel) Enter(v *Visitor) {
	assert.IsTrue(v != nil)

	if v.IsDebug {
		assert.IsTrue(t.replayId == 0)
	} else {
		assert.IsTrue(t.replayId != 0)
		if v.ReplayId == 0 {
			v.ReplayId = t.replayId
		}

	}

	assert.IsTrue(v.ReplayId != 0)

	val, err := json.Marshal(v)
	if err != nil {
		println(err.Error())
	}
	println(string(val))
	t.actionLogger.InsertAction(v.ReplayId, v.MessageId, string(v.ActionName), v.CausedBy, string(v.ActionType), string(v.ActionDirection), v.Payload, string(v.ActionType))
	t.queue <- v
}

func (t *Tunnel) NextVisitor() (*Visitor, error) {
	v := <-t.queue

	switch v.ActionType {
	case INPUT:
		switch v.ActionName {
		case TICK:
			println("Not doing anything for the tick...")
			return nil, nil
		case LOGON:
			break
		default:
			panic("unrecognized actionName: " + v.ActionName)
		}
	case REQUEST:
		panic("There are no request topics yet...")
	case REPLY:
		panic("There are no reply topics yet...")
	default:
		panic("unrecognized actionDirection")
	}
	return v, nil
}

func (t *Tunnel) Exit(v *Visitor) (*Visitor, error) {
	// Only record if we have a valid replay ID
	if v.ReplayId == 0 {
		log.Printf("WARNING: Skipping RecordOut - visitor has no replay ID (message: %s)", v.MessageId)
		return nil, fmt.Errorf("RecordOut called with nil visitor")
	}
	val, err := json.Marshal(v)
	if err != nil {
		println(err.Error())
	}
	println(string(val))
	t.actionLogger.InsertAction(v.ReplayId, v.MessageId, string(v.ActionName), v.CausedBy, string(v.ActionType), string(v.ActionDirection), v.Payload, string(v.ActionType))
	return v, nil
}

func NewInputAction(topic ActionName, payload string) *Visitor {
	return &Visitor{
		ActionDirection: IN,
		ActionType:      INPUT,
		MessageId:       generateMessageId(),
		ActionName:      topic,
		CausedBy:        "M0",
		Payload:         payload,
		IsDebug:         false,
		ReplayId:        0, // Will be set by Tunnel when enqueued
	}
}

func NewVisitorFromActionRow(messageID, topic, causedBy, messageType, direction, payload string, replayID int64) *Visitor {
	return &Visitor{
		MessageId:       messageID,
		ActionName:      ActionName(topic),
		CausedBy:        causedBy,
		Payload:         payload,
		ActionType:      ActionType(messageType),
		ActionDirection: ActionDirection(direction),
		IsDebug:         true,
		ReplayId:        replayID,
	}
}

func NewNormalTunnel(actionLogger *ActionLogger) *Tunnel {
	fileName := generateFileName()
	assert.IsTrue(fileName != "")

	fileId, err := actionLogger.InsertReplay("Normal Run Created", fileName, 1, nil)
	if err != nil {
		assert.IsTrue(false) // This cannot happen and should not happen
	}
	replayId := fileId

	return &Tunnel{
		queue:        make(chan *Visitor, 100),
		replayId:     replayId,
		actionLogger: actionLogger,
	}
}

func NewDebugTunnel(actionLogger *ActionLogger) *Tunnel {
	return &Tunnel{
		queue:        make(chan *Visitor, 100),
		replayId:     0,
		actionLogger: actionLogger,
	}
}

func generateMessageId() string {
	charset := "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	messageId := "M"
	for i := 0; i < 5; i++ {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			return ""
		}
		messageId = messageId + string(charset[n.Int64()])
	}
	return messageId
}

func generateFileName() string {
	charset := "abcdefghjklmnpqrstxyz"
	fileId := ""
	for i := 0; i < 7; i++ {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			return ""
		}
		fileId = fileId + string(charset[n.Int64()])
	}
	return fileId
}
