// Package bot_interfaces_http_ws is a websocket handler for the http mux
package bot_interfaces_http_ws

import (
	"connectly-interview/internal/bot/domain/bot_chat"
	"fmt"
	"github.com/google/uuid"
	"golang.org/x/net/websocket"
	"log"
	"net/http"
	"reflect"
)

type Websockets struct {
	ws                    *websocket.Server
	sendChan              chan []byte
	receiveChan           chan<- []byte
	newChatHandler        func()
	newChatMessageHandler func(chatId bot_chat.ChatId, msg []byte) error
}

type Args struct {
	ReceiveChan           chan<- []byte
	NewChatHandler        func()
	NewChatMessageHandler func(chatId bot_chat.ChatId, msg []byte) error
}

func New(args Args) *Websockets {
	w := &Websockets{}

	wsServer := &websocket.Server{}
	w.ws = wsServer

	sendChan := make(chan []byte, 128) // 128 to make hiccups "less visible"
	w.sendChan = sendChan
	w.receiveChan = args.ReceiveChan

	return w
}

type MsgType uint8

const (
	MsgTypeOutcomingReply = iota + 1
	MsgTypeIncomingReply
	MsgTypeNewChat
)

type IncomingReply struct {
	ChatId bot_chat.ChatId `json:"chat_id"`
	Reply  []byte          `json:"reply"`
}

type NewChatMsg struct {
	Username string
}

type OutMsg struct {
	MsgType MsgType `json:"msg_type"`
	Msg     []byte  `json:"msg"`
}

func (msg *OutMsg) IsEmpty() bool {
	return msg == nil || msg.MsgType == 0 || len(msg.Msg) == 0
}

type InMsg struct {
	MsgType MsgType     `json:"msg_type"`
	Msg     interface{} `json:"msg"`
}

func (msg *InMsg) IsEmpty() bool {
	return msg == nil || msg.MsgType == 0 || msg.Msg == nil
}

func (websockets *Websockets) Handler(w http.ResponseWriter, r *http.Request) {
	if websockets == nil {
		panic("websockets is nil")
	}

	s := websocket.Server{Handler: websocket.Handler(websockets.handleWebSocket)}
	s.ServeHTTP(w, r)
}

func (websockets *Websockets) handleWebSocket(ws *websocket.Conn) {
	if websockets == nil {
		panic("websockets is nil")
	}

	defer ws.Close()
	fmt.Printf("debug: opened ws connection\n")
	go func() {
		for {
			select {
			case r := <-websockets.sendChan:
				err := websocket.Message.Send(ws, r)
				if err != nil {
					return
				}
			}
		}
	}()
	for {
		var err error
		var inMsg InMsg
		if err = websocket.JSON.Receive(ws, &inMsg); err != nil {
			if err2 := websocket.Message.Send(ws, `{ "msgType": 3 "msg": "error: "`+err.Error()+`"}'`); err2 != nil {
				log.Printf("incoming message: %q,\ncan't send error: %s\n", err2, err)
			}
			continue
		}

		outMsg, err := websockets.processMessage(inMsg)
		if err != nil {
			fmt.Printf("could not process websocket message: %q\n", err)
		}

		// Send a response back - if any
		if outMsg.IsEmpty() {
			continue
		}

		if err := websocket.JSON.Send(ws, outMsg); err != nil {
			log.Println("can't send: ", err)
			if err2 := websocket.Message.Send(ws, `{ "msgType": 3 "msg": "error: "`+err.Error()+`"}'`); err2 != nil {
				log.Printf("incoming message: %q,\ncan't send error: %s\n", err2, err)
			}
			continue
		}
	}
}

func (websockets *Websockets) processMessage(inMsg InMsg) (*OutMsg, error) {
	if websockets == nil {
		panic("websockets is nil")
	}

	switch inMsg.Msg.(type) {
	case map[string]interface{}:
		break
	default:
		return nil, fmt.Errorf("inMsg.msg is not a valid type")
	}
	//var err error
	switch inMsg.MsgType {
	case MsgTypeIncomingReply:
		fmt.Printf("%+v,\n\n%q\ntype: %v\n", inMsg.Msg, inMsg, reflect.TypeOf(inMsg.Msg))
		var jsonMsg IncomingReply

		msg := inMsg.Msg.(map[string]interface{})

		// get chat uuid
		chatId := msg["chat_id"].(string)
		chatUuid, err := uuid.Parse(chatId)
		if err != nil {
			return nil, fmt.Errorf("invalid chat uuid, not a proper uuid")
		}

		// get reply
		reply := msg["reply"].(string)

		// assign to json message
		jsonMsg.ChatId = bot_chat.ChatId(chatUuid)
		jsonMsg.Reply = []byte(reply)

		// send to handler
		if websockets == nil {
			return nil, fmt.Errorf("wtf, websockets is nil?")
		}
		if websockets.newChatMessageHandler == nil {
			return nil, fmt.Errorf("no chat message handler provided")
		}

		err = websockets.newChatMessageHandler(jsonMsg.ChatId, jsonMsg.Reply)
		if err != nil {
			return nil, fmt.Errorf("could not process new chat message: %w", err)
		}
	case MsgTypeNewChat:
		websockets.newChatHandler()
	}
	// Implement your request processing logic here
	return nil, nil
}
