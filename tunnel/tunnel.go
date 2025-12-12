package tunnel

import (
	"crypto/rand"
	"fmt"
	"fund78/action_logger"
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
	queue    []*visitor
	replayId int64
	explorer *action_logger.ActionLogger
}

type visitor struct {
	messageId   string
	topic       ActionName
	causedBy    string
	payload     string
	messageType ActionType
	direction   ActionDirection
	actionType  string
	replayId    int64
}

func (t *Tunnel) Enter(v *visitor) {
	assert.IsTrue(v != nil)

	if v.replayId == 0 {
		v.replayId = t.replayId
	}

	assert.IsTrue(v.replayId != 0)

	t.explorer.InsertAction(v.replayId, v.messageId, string(v.topic), v.causedBy, string(v.messageType), string(v.direction), v.payload, v.actionType)
	t.queue = append(t.queue, v)
}

func (t *Tunnel) NextVisitor() (*visitor, error) {
	if len(t.queue) == 0 {
		return nil, fmt.Errorf("Queue is empty")
	}
	v := (t.queue)[0]
	(t.queue)[0] = nil // Optional: zero out the element
	t.queue = (t.queue)[1:]

	topic := GetTopic(v)
	messageType := GetMessageType(v)
	switch messageType {
	case INPUT:
		switch topic {
		case TICK:
			break
		case LOGON:
			break
		default:
			panic("unrecognized topic: " + topic)
		}
	case REQUEST:
		panic("There are no request topics yet...")
	case REPLY:
		panic("There are no reply topics yet...")
	default:
		panic("unrecognized direction")
	}
	return v, nil
}

func (t *Tunnel) Exit(v *visitor) (*visitor, error) {
	// Only record if we have a valid replay ID
	if v.replayId == 0 {
		log.Printf("WARNING: Skipping RecordOut - visitor has no replay ID (message: %s)", v.messageId)
		return nil, fmt.Errorf("RecordOut called with nil visitor")
	}
	t.explorer.InsertAction(v.replayId, v.messageId, string(v.topic), v.causedBy, string(v.messageType), string(v.direction), v.payload, v.actionType)
	return v, nil
}

func NewInputAction(topic ActionName, payload string) *visitor {
	return &visitor{
		direction:   IN,
		messageType: INPUT,
		messageId:   generateMessageId(),
		topic:       topic,
		causedBy:    "M0",
		payload:     payload,
		actionType:  "r",
		replayId:    0, // Will be set by Tunnel when enqueued
	}
}

func NewVisitorFromActionRow(messageID, topic, causedBy, messageType, direction, payload string, replayID int64) *visitor {
	return &visitor{
		messageId:   messageID,
		topic:       ActionName(topic),
		causedBy:    causedBy,
		payload:     payload,
		messageType: ActionType(messageType),
		direction:   ActionDirection(direction),
		actionType:  "d",
		replayId:    replayID,
	}
}

func GetTopic(v *visitor) ActionName {
	return v.topic
}

func GetMessageType(v *visitor) ActionType {
	return v.messageType
}

func NewNormalTunnel() *Tunnel {
	actionLogger := action_logger.NewActionLogger()

	fileName := generateFileName()
	assert.IsTrue(fileName != "")

	fileId, err := actionLogger.InsertReplay("Normal Run Created", fileName, 1, nil)
	if err != nil {
		assert.IsTrue(false) // This cannot happen and should not happen
	}
	replayId := fileId

	return &Tunnel{
		queue:    make([]*visitor, 0),
		replayId: replayId,
		explorer: actionLogger,
	}
}

func NewDebugTunnel() *Tunnel {
	actionLogger := action_logger.NewActionLogger()

	return &Tunnel{
		queue:    make([]*visitor, 0),
		replayId: 0,
		explorer: actionLogger,
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
