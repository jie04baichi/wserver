package wserver

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/gorilla/websocket"
)

const (
	serverDefaultWSPath   = "/ws"
	serverDefaultSendPath = "/send"
)

var defaultUpgrader = &websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(*http.Request) bool {
		return true
	},
}

// Server defines parameters for running websocket server.
type Server struct {
	// Path for websocket request, default "/ws".
	WSPath string

	// Path for send message, default "/send".
	SendPath string

	// Upgrader is for upgrade connection to websocket connection using
	// "github.com/gorilla/websocket".
	//
	// If Upgrader is nil, default upgrader will be used. Default upgrader is
	// set ReadBufferSize and WriteBufferSize to 1024, and CheckOrigin always
	// returns true.
	Upgrader *websocket.Upgrader
}

// Listen listens on the TCP network address addr.
func (s *Server) Listen(addr string) error {
	b := &binder{
		userID2EventConnMap: make(map[string]*[]eventConn),
		connID2UserIDMap:    make(map[string]string),
	}

	// websocket request handler
	wh := websocketHandler{
		upgrader: s.Upgrader,
		binder:   b,
	}
	if wh.upgrader == nil {
		wh.upgrader = defaultUpgrader
	}
	http.Handle(s.WSPath, &wh)

	// send request handler
	sh := sendHandler{
		binder: b,
	}
	http.Handle(s.SendPath, &sh)

	return http.ListenAndServe(addr, nil)
}

// Check parameters of Server, returns error if fail.
func (s Server) check() error {
	if !checkPath(s.WSPath) {
		return fmt.Errorf("WSPath: %s not illegal", s.WSPath)
	}
	if !checkPath(s.SendPath) {
		return fmt.Errorf("SendPath: %s not illegal", s.SendPath)
	}
	if s.WSPath == s.SendPath {
		return errors.New("WSPath is equal to SendPath")
	}

	return nil
}

// NewServer creates a new Server.
//
// WSPath and SendPath can't be empty.
func NewServer(WSPath, SendPath string) *Server {
	return &Server{
		WSPath:   WSPath,
		SendPath: SendPath,
	}
}

func checkPath(path string) bool {
	if path != "" && !strings.HasPrefix(path, "/") {
		return false
	}
	return true
}
