package model

// Channel is the service whereby servers exchange info
type Channel struct {

	// MsgStats is the content of the stats reported by the Ethereum node
	MsgPing    chan []byte
	MsgLatency chan []byte

	// Nodes registered to the relay server
	Nodes map[string][]byte

	//use for flag the login client
	LoginIDs map[string]string
}
