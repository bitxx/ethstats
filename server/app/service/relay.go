package service

import (
	"encoding/json"
	"ethstats/common/util/connutil"
	"ethstats/common/util/emailutil"
	"ethstats/server/app/model"
	"ethstats/server/config"
	"fmt"
	"github.com/bitxx/logger/logbase"
	"github.com/gorilla/websocket"
	"net/http"
	"time"
)

const (
	messageHello   string = "hello"
	messagePing    string = "node-ping"
	messageLatency string = "latency"
	messageStats   string = "stats"
)

// NodeRelay contains the secret used to authenticate the communication between
// the Ethereum node and this server
type NodeRelay struct {
	secret                string
	logger                *logbase.Helper
	channel               *model.Channel
	emailLastNodeErrCache map[string]*time.Time
}

// NewRelay creates a new NodeRelay struct with required fields
func NewRelay(channel *model.Channel, logger *logbase.Helper) *NodeRelay {
	return &NodeRelay{
		channel: channel,
		secret:  config.ApplicationConfig.Secret,
		logger:  logger,
		//emailLastNodeErrCache: make(map[string]*time.Time),
	}
}

// Close closes the connection between this server and all Ethereum nodes connected to it
func (n *NodeRelay) Close() {
	close(n.channel.MsgPing)
	close(n.channel.MsgLatency)
}

// HandleRequest is the function to handle all server requests that came from
// Ethereum nodes
func (n *NodeRelay) HandleRequest(w http.ResponseWriter, r *http.Request) {
	upgradeConn := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
	conn, err := connutil.NewUpgradeConn(upgradeConn, w, r)
	if err != nil {
		n.logger.Warnf("error establishing node connection: %s", err)
		return
	}
	n.logger.Infof("new node connected! (addr=%s, host=%s)", r.RemoteAddr, r.Host)
	go n.loop(conn)
}

// loop loops as long as the connection is alive and retrieves node packages
func (n *NodeRelay) loop(c *connutil.ConnWrapper) {
	errType := 0 //1-ping error 2-node stopped
	// Close connection if an unexpected error occurs and delete the node
	// from the map of connected nodes...
	defer func(c *connutil.ConnWrapper) {
		//remove the error node and login id\
		if n.channel.Nodes[c.RemoteAddr().String()] != nil {
			delete(n.channel.Nodes, c.RemoteAddr().String())
		}

		//send email
		if n.channel.LoginIDs[c.RemoteAddr().String()] != "" {
			nodeErrLastTime := n.emailLastNodeErrCache[n.channel.LoginIDs[c.RemoteAddr().String()]]
			//cache time for delay send the same node email
			now := time.Now()
			if nodeErrLastTime == nil || (nodeErrLastTime != nil && now.Sub(*nodeErrLastTime).Hours() > 1) {
				n.emailLastNodeErrCache[n.channel.LoginIDs[c.RemoteAddr().String()]] = &now
				content := ""
				switch errType {
				case 1:
					//ping error
					content = "node: [" + n.channel.LoginIDs[c.RemoteAddr().String()] + "-" + c.RemoteAddr().String() + "] ping error"
				case 2:
					//node stopped
					content = "node: [" + n.channel.LoginIDs[c.RemoteAddr().String()] + "-" + c.RemoteAddr().String() + "] process stopped"
				default:
					return
				}

				err := emailutil.SendEmailDefault(fmt.Sprintf("%s-node error\n", time.Now().Format("2006-01-02 15:04:05")), content)
				if err != nil {
					n.logger.Error("email content: ", content, " send error: ", err)
				} else {
					n.logger.Info("email send success")
				}
			}

			//remove error node
			delete(n.channel.LoginIDs, c.RemoteAddr().String())
		}
		err := c.Close()
		if err != nil {
			n.logger.Error(err)
			return
		}
		n.logger.Warnf("connection with node closed, there are %d connected nodes", len(n.channel.Nodes))
	}(c)
	pingErrCount := 0 // if ping error count big than 100, it means node is error,maybe node process stopped or network error
	// Client loop
	for {
		_, content, err := c.ReadMessage()
		if err != nil {
			n.logger.Errorf("error reading message from client, %s", err)
			return
		}
		// Create emitted message from the node
		msg := model.Message{Content: content}
		msgType, err := msg.GetType()
		if err != nil {
			n.logger.Warnf("can't get type of message sent by the node: %s", err)
			return
		}
		switch msgType {
		case messageHello:
			authMsg, parseError := parseAuthMessage(msg)
			if parseError != nil {
				n.logger.Warnf("can't parse authorization message sent by node[%s], error: %s", authMsg.ID, parseError)
				loginErr := authMsg.SendLoginErrResponse(c, "login data parsing error")
				if loginErr != nil {
					n.logger.Errorf("error sending authorization response [parse message error info] to node[%s], error: %s", authMsg.ID, loginErr)
					return
				}
				return
			}
			// first check if the secret is correct
			if authMsg.Secret != n.secret {
				n.logger.Errorf("invalid secret from node %s, can't get stats", authMsg.ID)
				loginErr := authMsg.SendLoginErrResponse(c, "authorization error,invalid secret")
				if loginErr != nil {
					n.logger.Errorf("error sending authorization response [invalid secret] to node[%s], error: %s", authMsg.ID, loginErr)
					return
				}
				return
			}
			//判断节点名称是否重复，遍历效率有点低，有时间了在考虑怎么优化，或者伙计们可以帮忙想个简单的法子
			for k, v := range n.channel.LoginIDs {
				if v == authMsg.ID && k != c.RemoteAddr().String() {
					n.logger.Errorf("the id [%s] has login", authMsg.ID)
					loginErr := authMsg.SendLoginErrResponse(c, "the login id has being exist,please change the id name")
					if loginErr != nil {
						n.logger.Errorf("error sending authorization response [login id is exist] to node[%s], error: %s", authMsg.ID, loginErr)
						return
					}
					return
				}

			}
			sendError := authMsg.SendResponse(c)
			if sendError != nil {
				n.logger.Errorf("error sending authorization response to node[%s], error: %s", authMsg.ID, sendError)
				return
			}
			n.channel.LoginIDs[c.RemoteAddr().String()] = authMsg.ID
		case messagePing:
			// When the node emit a ping message, we need to respond with pong
			// before five seconds to authorize that node to sent reports
			ping, err := parseNodePingMessage(msg)
			if err != nil {
				pingErrCount++
				if pingErrCount >= 100 {
					errType = 1
					n.logger.Warnf("can't parse ping message sent by node[%s], error: %s", ping.ID, err)
					return
				}

				continue
			}
			if ping.NodeStatus == "stopped" {
				errType = 2
				n.logger.Warnf("node[%s] process stopped", ping.ID)
				return
			}
			sendError := ping.SendResponse(c)
			if sendError != nil {
				n.logger.Errorf("error sending pong response to node[%s], error: %s", ping.ID, sendError)
			}
			n.channel.MsgPing <- content
		case messageLatency:
			n.channel.MsgLatency <- content
		case messageStats:
			// use node addr as identifier to check node availability
			n.channel.Nodes[c.RemoteAddr().String()] = content
			n.logger.Infof("currently there are %d connected nodes", len(n.channel.Nodes))
		}
	}
}

// parseNodePingMessage parse the current ping message sent bu the Ethereum node
// and creates a message.NodePing struct with that info
func parseNodePingMessage(msg model.Message) (*model.NodePing, error) {
	value, err := msg.GetValue()
	if err != nil {
		return &model.NodePing{}, err
	}
	var ping model.NodePing
	err = json.Unmarshal(value, &ping)
	return &ping, err
}

// parseAuthMessage parse the current byte array and transforms it to an AuthMessage struct.
// If an error occurs when json unmarshal, an error is returned
func parseAuthMessage(msg model.Message) (*model.AuthMessage, error) {
	value, err := msg.GetValue()
	if err != nil {
		return &model.AuthMessage{}, err
	}
	var detail model.AuthMessage
	err = json.Unmarshal(value, &detail)
	return &detail, err
}
