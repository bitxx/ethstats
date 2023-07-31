package model

import (
	"ethstats/common/util/connutil"
)

// AuthMessage is the struct sent by the server on the first connection
type AuthMessage struct {
	ID     string `json:"id"`
	Secret string `json:"secret"`
}

// SendResponse send the ready response to the node to initiate the communication
func (a *AuthMessage) SendResponse(c *connutil.ConnWrapper) error {
	return c.WriteJSON(map[string][]interface{}{"emit": {"ready"}})
}

func (a *AuthMessage) SendLoginErrResponse(c *connutil.ConnWrapper, errMsg string) error {
	return c.WriteJSON(map[string][]interface{}{"emit": {"un-authorization", errMsg}})
}
