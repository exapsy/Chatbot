package bot_interface_http

import (
	"connectly-interview/internal/bot/domain/bot_chat"
	bot_interfaces_http_ws "connectly-interview/internal/bot/interfaces/http_server/ws"
	"context"
	"fmt"
	"net/http"
)

func NewErrRunBotServer(err error) *ErrRunBotServer {
	return &ErrRunBotServer{
		err: err,
	}
}

type ErrRunBotServer struct {
	err error
}

func (e ErrRunBotServer) Error() string {
	return "could not run bot http_server"
}

type Server interface {
	Listen() (<-chan []byte, <-chan error)
}

type server struct {
	http_server        *http.Server
	ctx                context.Context
	certFile           *string
	keyFile            *string
	address            string
	ws                 *bot_interfaces_http_ws.Websockets
	wsIncomingMsgsChan chan []byte
}

type Server_Args struct {
	Context               context.Context
	Address               string
	CertFile              *string
	KeyFile               *string
	ExtraMiddlewares      []http.Handler
	NewChatHandler        func()
	NewChatMessageHandler func(chatId bot_chat.ChatId, msg []byte) error
}

func New(args Server_Args) Server {
	server := &server{}

	// region initiate vars
	m := http.NewServeMux()

	// region add endpoints

	if args.NewChatHandler == nil {
		panic("no chat handler provided to http server")
	}
	if args.NewChatMessageHandler == nil {
		panic("no chat message handler provided to http server")
	}

	wsReceiveChan := make(chan []byte)
	server.ws = bot_interfaces_http_ws.New(bot_interfaces_http_ws.Args{
		ReceiveChan:           wsReceiveChan,
		NewChatHandler:        args.NewChatHandler,
		NewChatMessageHandler: args.NewChatMessageHandler,
	})
	m.HandleFunc("/ws/", server.ws.Handler)

	// endregion

	wsIncomingMsgsChan := make(chan []byte)
	ws := bot_interfaces_http_ws.New(bot_interfaces_http_ws.Args{
		ReceiveChan: wsIncomingMsgsChan,
	})
	http_server := &http.Server{
		Addr:    args.Address,
		Handler: m,
	}

	// endregion

	// region initiate server values

	server.http_server = http_server
	server.ctx = args.Context
	server.certFile = args.CertFile
	server.keyFile = args.KeyFile
	server.address = args.Address
	server.ws = ws
	server.wsIncomingMsgsChan = wsIncomingMsgsChan

	// endregion

	return server
}

func (s *server) Listen() (reqsChan <-chan []byte, errChan <-chan error) {
	errChan = make(chan error)
	reqsChan = make(chan []byte)
	var err error

	go func() {
		if s.certFile != nil && s.keyFile != nil {
			fmt.Printf("listening http tls server at %q\n", s.address)
			err = s.http_server.ListenAndServeTLS(*s.certFile, *s.keyFile)
			if err != nil {
				panic(NewErrRunBotServer(err))
				return
			}
		} else {
			fmt.Printf("listening http server at %q\n", s.address)
			err = s.http_server.ListenAndServe()
			if err != nil {
				panic(NewErrRunBotServer(err))
				return
			}
		}

		for {
			select {
			case <-s.ctx.Done():
				err := s.http_server.Shutdown(s.ctx)
				if err != nil {
					panic("error shutting down bot http_server")
					return
				}
			}
		}
	}()

	return reqsChan, errChan
}
