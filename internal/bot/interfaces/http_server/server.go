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

type Server struct {
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

func New(args Server_Args) *Server {
	m := http.NewServeMux()
	server := &http.Server{
		Addr:    args.Address,
		Handler: m,
	}
	return &Server{
		http_server: server,
		ctx:         args.Context,
		certFile:    args.CertFile,
		keyFile:     args.KeyFile,
		address:     args.Address,
	}
}

func (s *Server) Start() <-chan error {
	var err error

	go func() {
		if s.certFile != nil && s.keyFile != nil {
			err = s.http_server.ListenAndServeTLS(*s.certFile, *s.keyFile)
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

	return nil
}

func (s *Server) Listen() <-chan string {
	panic("not implemented")
	return nil
}
