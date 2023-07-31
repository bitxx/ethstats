package model

import (
	"ethstats/common/util/connutil"
)

// NodePing contains the last time the node is alive
type NodePing struct {
	ID   string `json:"id"`
	Time string `json:"clientTime"`
}

// SendResponse send the pong response to the node
func (n *NodePing) SendResponse(c *connutil.ConnWrapper) error {
	// message type is always 1
	return c.WriteJSON(map[string][]interface{}{"emit": {"node-pong", n.ID}})
}
