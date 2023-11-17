// Package bot_interfaces_http_ws is a websocket handler for the http mux
package bot_interfaces_http_ws

import "golang.org/x/net/websocket"

type Websockets struct {
	ws *websocket.Server
}

func New(addr string) *Websockets {
	w := &Websockets{}
	wsServer := &websocket.Server{}

	return w
}

type MsgType uint8

const (
	MsgTypeOutcomingReply = iota + 1
	MsgTypeIncomingReply
)

type OutMsg struct {
	MsgType MsgType `json:"msg_type"`
	Msg     []byte  `json:"msg"`
}

type InMsg struct {
	MsgType MsgType `json:"msg_type"`
	Msg     []byte  `json:"msg"`
}

func (w *Websockets) Listen() (reqs <-chan []byte, errors <-chan error) {

	return nil, nil
}
