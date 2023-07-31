package app

import (
	"context"
	"encoding/json"
	"errors"
	"ethstats/client/app/model"
	"ethstats/client/config"
	"ethstats/common/util/connutil"
	"github.com/bitxx/ethutil"
	"github.com/bitxx/logger/logbase"
	"os"
	"os/signal"
	"runtime"
	"strconv"
	"strings"
	"time"
)

type App struct {
	node    model.Node
	readyCh chan struct{}
	pongCh  chan struct{}
	logger  *logbase.Helper
}

func NewApp() *App {
	node := model.Node{
		Id:         config.ApplicationConfig.Name,
		Name:       config.ApplicationConfig.Name,
		Contact:    config.ApplicationConfig.Contract,
		ChainPort:  config.ChainConfig.Port,
		OSPlatform: runtime.GOARCH,
		OS:         runtime.GOOS,
		Client:     config.ApplicationConfig.Version,
	}

	return &App{
		node:    node,
		readyCh: make(chan struct{}),
		pongCh:  make(chan struct{}),
		logger:  logbase.NewHelper(logbase.DefaultLogger),
	}
}

func (a *App) Start() {
	// logbase.NewHelper(core.Runtime.GetLogger())
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	var err error
	isInterrupt := false

	conn := &connutil.ConnWrapper{}
	readTicker := time.NewTimer(0)
	latencyTicker := time.NewTimer(0)

	defer func() {
		a.close(conn, readTicker, latencyTicker)
		// if not interrupt,restart the client
		if !isInterrupt {
			time.Sleep(5 * time.Second)
			a.Start()
		}
	}()

	conn, err = connutil.NewDialConn(config.ApplicationConfig.ServerUrl)
	if err != nil {
		a.logger.Warn("dial error: ", err)
		return
	}

	for {
		select {
		case <-readTicker.C:
			//after customer,change the time
			readTicker.Reset(10 * time.Second)
			latencyTicker.Reset(2 * time.Second)

			//login
			login := map[string][]interface{}{
				"emit": {"hello", map[string]string{
					"id":     a.node.Name,
					"secret": config.ApplicationConfig.Secret,
				}},
			}
			err = conn.WriteJSON(login)
			if err != nil {
				return
			}

			//read info
			go a.readLoop(conn)

			select {
			case <-latencyTicker.C:
				if err = a.reportLatency(conn); err != nil {
					a.logger.Warn("requested latency report failed: ", err)
				}
			case <-a.readyCh:
				//登录成功，上传数据
				if err = a.reportStats(conn); err != nil {
					a.logger.Warn("stats info report failed: ", err)
				}
			}
		case <-interrupt:
			a.close(conn, readTicker, latencyTicker)
			isInterrupt = true
			return
		}
	}
}

func (a *App) readLoop(conn *connutil.ConnWrapper) {
	for {
		blob := json.RawMessage{}
		if err := conn.ReadJSON(&blob); err != nil {
			a.logger.Warn("received and decode message error: ", err)
			return
		}
		// If the network packet is a system ping, respond to it directly
		var ping string
		if err := json.Unmarshal(blob, &ping); err == nil && strings.HasPrefix(ping, "primus::ping::") {
			if err := conn.WriteJSON(strings.Replace(ping, "ping", "pong", -1)); err != nil {
				a.logger.Warn("failed to respond to system ping message: ", err)
				return
			}
			continue
		}
		// Not a system ping, try to decode an actual state message
		var msg map[string][]interface{}
		if err := json.Unmarshal(blob, &msg); err != nil {
			a.logger.Warn("failed to decode message: ", err)
			return
		}

		if len(msg["emit"]) == 0 {
			a.logger.Warn("received message invalid: ", msg)
			return
		}
		msgType, ok := msg["emit"][0].(string)
		if !ok {
			a.logger.Warn("received invalid message type: ", msg["emit"][0])
			return
		}
		a.logger.Trace("received message type: ", msgType)

		switch msgType {
		case "ready":
			//只有接收到了ready信息，才初始化获取数据
			a.logger.Info("connect success!")
			a.readyCh <- struct{}{}
		case "un-authorization":
			if len(msg["emit"]) >= 2 {
				if errMsg, ok := msg["emit"][1].(string); ok {
					a.logger.Warn(errMsg)
				}
			}
			return
		case "node-pong":
			a.pongCh <- struct{}{}
		}

	}
}

func (a *App) reportLatency(conn *connutil.ConnWrapper) error {
	start := time.Now()

	ping := map[string][]interface{}{
		"emit": {"node-ping", map[string]string{
			"id":         config.ApplicationConfig.Name,
			"clientTime": start.String(),
		}},
	}

	if err := conn.WriteJSON(ping); err != nil {
		return err
	}
	// Wait for the pong request to arrive back
	select {
	case <-a.pongCh:
		// Pong delivered, report the latency
	case <-time.After(10 * time.Second):
		// MsgPing timeout, abort
		return errors.New("ping timed out")
	}
	latency := strconv.Itoa(int((time.Since(start) / time.Duration(2)).Nanoseconds() / 1000000))

	// Send back the measured latency
	a.logger.Trace("sending measured latency: ", latency)

	stats := map[string][]interface{}{
		"emit": {"latency", map[string]string{
			"id":      config.ApplicationConfig.Name,
			"latency": latency,
		}},
	}
	return conn.WriteJSON(stats)
}

func (a *App) reportStats(conn *connutil.ConnWrapper) error {
	ethClient := ethutil.NewEthClient(config.ChainConfig.Url, config.ChainConfig.Timeout)
	chain, err := ethClient.Chain()
	if err != nil {
		return err
	}
	c := chain.RemoteRpcClient
	// peer count
	peerCount, _ := c.PeerCount(context.Background())

	// is active
	active := false
	if peerCount > 0 {
		active = true
	}

	// gas price
	gasPrice, _ := c.SuggestGasPrice(context.Background())

	// is syncing
	process, err := c.SyncProgress(context.Background())
	syncing := false
	if err == nil && process != nil {
		progress := process.CurrentBlock - process.StartingBlock
		total := process.HighestBlock - process.StartingBlock
		if progress/total < 1 {
			syncing = true
		}
	}

	// latest block
	latestBlock, err := c.BlockByNumber(context.Background(), nil)
	block := model.Block{}
	if err == nil {
		block.Number = latestBlock.NumberU64()
		block.Hash = latestBlock.Hash().String()
		block.Difficulty = latestBlock.Difficulty().Uint64()
		block.Time = latestBlock.Time()
		//block.Transactions = latestBlock.Transactions()
		//block.Uncles = latestBlock.Uncles()
	}
	pendingCount, _ := c.PendingTransactionCount(context.Background())

	stats := model.Stats{
		NodeInfo:  a.node,
		Active:    active,
		PeerCount: peerCount,
		Pending:   pendingCount,
		GasPrice:  gasPrice.Int64(),
		Syncing:   syncing,
		Block:     &block,
	}
	report := map[string][]interface{}{
		"emit": {"stats", stats},
	}
	return conn.WriteJSON(report)
}

func (a *App) close(conn *connutil.ConnWrapper, readTicker, latencyTicker *time.Timer) {
	if conn != nil {
		_ = conn.Close()
	}
	if readTicker != nil {
		_ = readTicker.Stop()
	}
	if latencyTicker != nil {
		_ = latencyTicker.Stop()
	}
}
