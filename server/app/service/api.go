package service

import (
	"ethstats/common/util/connutil"
	"ethstats/common/util/emailutil"
	"ethstats/server/app/model"
	"ethstats/server/config"
	"fmt"
	"github.com/bitxx/logger/logbase"
	"github.com/gorilla/websocket"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// Api is the responsible to send node state to registered hub
type Api struct {
	logger *logbase.Helper
	hub    *hub
}

// NewApi creates a new Api struct with the required service
func NewApi(channel *model.Channel, logger *logbase.Helper) *Api {
	hub := &hub{
		register: make(chan *connutil.ConnWrapper),
		logger:   logger,
		close:    make(chan interface{}),
		clients:  make(map[*connutil.ConnWrapper]bool),
		channel:  channel,
	}
	go hub.loop()
	return &Api{
		logger: logger,
		hub:    hub,
	}
}

// Close this server and all registered client connections
func (a *Api) Close() {
	a.logger.Info("prepared to close all client connections")
	a.hub.close <- "close"
}

// HandleRequest handle all request from hub that are not Ethereum nodes
func (a *Api) HandleRequest(w http.ResponseWriter, r *http.Request) {
	upgradeConn := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
	conn, err := connutil.NewUpgradeConn(upgradeConn, w, r)
	if err != nil {
		a.logger.Errorf("error trying to establish communication with client (addr=%s, host=%s, URI=%s), %s",
			r.RemoteAddr, r.Host, r.RequestURI, err)
		return
	}
	a.logger.Infof("connected new client! (host=%s)", r.Host)
	a.hub.register <- conn
}

// hub maintain a list of registered clients to send messages
type hub struct {
	register chan *connutil.ConnWrapper
	logger   *logbase.Helper
	close    chan interface{}
	clients  map[*connutil.ConnWrapper]bool
	channel  *model.Channel
}

// loop loops as the server is alive and send messages to registered clients
func (h *hub) loop() {
	nodesReportTicker := time.NewTicker(15 * time.Second)
	nodesMonitorTicker := time.NewTicker(time.Duration(config.EmailConfig.MonitorTime) * time.Second)
	defer func() {
		nodesReportTicker.Stop()
		nodesMonitorTicker.Stop()
	}()
	//nodesMonitorTicker := time.NewTicker(1440 * time.Second)
	for {
		select {
		case client := <-h.register:
			h.clients[client] = true
		case ping := <-h.channel.MsgPing:
			//debug log for show the ping
			//h.logger.Info("debug log show ping = > ", string(ping))
			//use for send to any fronted client
			h.writeMessage(ping)
		case latency := <-h.channel.MsgLatency:
			//debug log for show the latency
			//h.logger.Info("debug log show latency = > ", string(latency))
			//use for send to any fronted client
			h.writeMessage(latency)
		case <-nodesReportTicker.C:
			if len(h.clients) <= 0 {
				continue
			}
			for _, v := range h.channel.Nodes {
				//debug log for show the node info
				//h.logger.Info("debug log show node = > ", string(v))
				//use for send to any fronted client
				h.writeMessage(v)
			}
		case <-nodesMonitorTicker.C:
			nodeCount := len(h.channel.Nodes)
			nodeInfo := ""
			for _, v := range h.channel.Nodes {
				//node info like this:
				//{"Active":true,"PeerCount":500,"Pending":94,"GasPrice":1097000000000,"Syncing":false,"NodeInfo":{"Id":"test","Name":"test","Contact":"","Coinbase":"","Node":"","Net":"","Protocol":"","Api":"","ChainPort":"30303","OSPlatform":"amd64","OS":"darwin","Client":"v1.0.0"},"Block":{"Number":484645,"Hash":"0x0036faa4d9b83ec836ebbcd6a98699323c41bdcdedf8d96a21c6dd25b9c41b88","Difficulty":0,"Transactions":null,"Uncles":null,"Time":1690774181}}

				idRe := regexp.MustCompile("Id\\\":(.*?),")
				idTmp1 := idRe.FindString(string(v))
				idTmp2 := strings.Replace(idTmp1, "Id\":\"", "", -1)
				id := strings.Replace(idTmp2, "\",", "", -1)

				re := regexp.MustCompile("Number\\\":(.*?),")
				heightTmp1 := re.FindString(string(v))
				heightTmp2 := strings.Replace(heightTmp1, "Number\":", "", -1)
				latestHigh := strings.Replace(heightTmp2, ",", "", -1)

				//ip := strings.Replace(k, "[::1]", "127.0.0.1", -1)
				nodeInfo = nodeInfo + "--节点ID：" + id + "，块高度：" + latestHigh + "\n"
			}
			content := "节点数量：" + strconv.Itoa(nodeCount) + "\n各节点块高度：\n" + nodeInfo
			fmt.Println(content)
			_ = emailutil.SendEmailDefault(fmt.Sprintf("%s-节点监控简报\n", time.Now().Format("2006-01-02 15:04:05")), content)
		case <-h.close:
			h.quit()
			break
		}
	}
}

// writeMessage to all registered clients. If an error occurs sending a message to a client,
// then these connection is closed and removed from the pool of registered clients
func (h *hub) writeMessage(msg []byte) {
	for client := range h.clients {
		err := client.WriteMessage(1, msg)
		if err != nil {
			h.logger.Infof("Closed connection with client: %s", client.RemoteAddr())
			// close and delete the client connection and release
			client.Close()
			delete(h.clients, client)
		}
	}
}

func (h *hub) quit() {
	h.logger.Info("Closing all registered clients")
	for client := range h.clients {
		client.Close()
		delete(h.clients, client)
	}
	close(h.register)
	close(h.close)
}
