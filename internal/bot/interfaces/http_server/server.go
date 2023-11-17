package bot_interface_http

import (
	"context"
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
	http_server *http.Server
	ctx         context.Context
	certFile    *string
	keyFile     *string
	address     string
}

type Server_Args struct {
	Context          context.Context
	Address          string
	CertFile         *string
	KeyFile          *string
	ExtraMiddlewares []http.Handler
}

func New(args Server_Args) Server {
	m := http.NewServeMux()
	s := &http.Server{
		Addr:    args.Address,
		Handler: m,
	}

	return &server{
		http_server: s,
		ctx:         args.Context,
		certFile:    args.CertFile,
		keyFile:     args.KeyFile,
		address:     args.Address,
	}
}

func newMux() *http.ServeMux {
	m := http.NewServeMux()
	m.HandleFunc("/ws/", func(w http.ResponseWriter, r *http.Request) {

	})
	return m
}

func (s *server) Listen() (reqsChan <-chan []byte, errChan <-chan error) {
	errChan = make(chan error)
	reqsChan = make(chan []byte)
	var err error

	go func() {
		if s.certFile != nil && s.keyFile != nil {
			err = s.http_server.ListenAndServeTLS(*s.certFile, *s.keyFile)
			if err != nil {
				panic(NewErrRunBotServer(err))
				return
			}
		} else {
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
