package model

import (
	"encoding/json"
)

// Message contains the Ethereum message
type Message struct {
	Content []byte
}

// GetType return the type of the message sent by the Ethereum node
func (e *Message) GetType() (string, error) {
	var content map[string][]interface{}
	err := json.Unmarshal([]byte(e.Content), &content)
	if err != nil {
		return "", err
	}
	result, _ := content["emit"][0].(string)
	return result, nil
}

// GetValue retrieve the current content of the emitted message by the node
func (e *Message) GetValue() ([]byte, error) {
	var content map[string][]interface{}
	err := json.Unmarshal([]byte(e.Content), &content)
	if err != nil {
		return nil, err
	}
	result, _ := content["emit"][1].(interface{})
	val, err := json.Marshal(result)
	return val, err
}
