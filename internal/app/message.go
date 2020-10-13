package app

import (
	"errors"
	"fmt"
)

type Message struct {
	Raw    JSON
	Action string
	Data   JSON
}

func NewMessage(action string, data JSON) *Message {
	return &Message{
		Raw:    nil,
		Action: action,
		Data:   data,
	}
}

func NewError(text string) *Message {
	return NewMessage("error", JSON{"text": text})
}

func GetString(key string, js JSON) string {
	d, e := js[key]
	if !e {
		return ""
	}
	switch v := d.(type) {
	case string:
		return v
	default:
		return ""
	}
}

func GetInt(key string, js JSON) int {
	d, e := js[key]
	if !e {
		return -1
	}
	switch v := d.(type) {
	case int:
		return v
	default:
		return -1
	}
}

func MessageFromJson(j JSON) (*Message, error) {
	action, exists := j["action"]
	if !exists {
		return nil, errors.New("action is missing")
	}

	data, _ := j["data"]

	dataJson := make(JSON)
	switch v := data.(type) {
	case map[string]interface{}:
		for s, a := range v {
			dataJson[s] = a
		}
	}

	return &Message{
		Raw:    j,
		Action: fmt.Sprintf("%s", action),
		Data:   dataJson,
	}, nil

}

type JSON map[string]interface{}

func (m *Message) Map() JSON {
	return JSON{
		"action": m.Action,
		"data":   m.Data,
	}
}
