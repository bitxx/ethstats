package connutil

import (
	"github.com/gorilla/websocket"
	"net"
	"net/http"
	"sync"
	"time"
)

type ConnWrapper struct {
	conn  *websocket.Conn
	rlock sync.Mutex
	wlock sync.Mutex
}

// NewDialConn 不加读写锁，执行时会出现问题
func NewDialConn(url string) (*ConnWrapper, error) {
	//发现有的节点因为网络问题，默认的45秒有点短，导致总timeout，这里先固定成120s，后续根据需要再考虑要不要可配置化
	dial := websocket.Dialer{
		Proxy:            http.ProxyFromEnvironment,
		HandshakeTimeout: 120 * time.Second,
	}
	c, _, err := dial.Dial(url, nil)
	if err != nil {
		return nil, err
	}
	return &ConnWrapper{conn: c}, nil
}

func NewUpgradeConn(upgradeConn websocket.Upgrader, w http.ResponseWriter, r *http.Request) (*ConnWrapper, error) {
	c, err := upgradeConn.Upgrade(w, r, nil)
	if err != nil {
		return nil, err
	}
	return &ConnWrapper{conn: c}, nil
}

// WriteJSON wraps corresponding method on the websocket but is safe for concurrent calling
func (w *ConnWrapper) WriteJSON(v interface{}) error {
	w.wlock.Lock()
	defer w.wlock.Unlock()

	return w.conn.WriteJSON(v)
}

// ReadJSON wraps corresponding method on the websocket but is safe for concurrent calling
func (w *ConnWrapper) ReadJSON(v interface{}) error {
	w.rlock.Lock()
	defer w.rlock.Unlock()
	return w.conn.ReadJSON(v)
}

func (w *ConnWrapper) WriteMessage(messageType int, data []byte) error {
	w.wlock.Lock()
	defer w.wlock.Unlock()

	return w.conn.WriteMessage(messageType, data)
}

func (w *ConnWrapper) ReadMessage() (messageType int, p []byte, err error) {
	w.rlock.Lock()
	defer w.rlock.Unlock()

	return w.conn.ReadMessage()
}

func (w *ConnWrapper) RemoteAddr() net.Addr {
	return w.conn.RemoteAddr()
}

// Close wraps corresponding method on the websocket but is safe for concurrent calling
func (w *ConnWrapper) Close() error {
	// The Close and WriteControl methods can be called concurrently with all other methods,
	// so the mutex is not used here
	return w.conn.Close()
}
