package bot_interfaces

import (
	"connectly-interview/internal/bot/interfaces/daemon"
	"connectly-interview/internal/bot/interfaces/http_server"
	"context"
	"fmt"
)

type Interface interface {
	// Start starts the communication interface
	Start() <-chan error
	// Listen listens to the prompts of the interface
	Listen() <-chan string
}

// TotalInterfaceTypes keeps track of how many interfaces there are
// Dangerous: it does not auto-update/generate // if the amount of interfaces changes, REMEMBER to change this
const TotalInterfaceTypes = 2

type InterfaceType uint8

const (
	InterfaceTypeHttpServer = iota + 1
	InterfaceTypeDaemon     = 2
)

type InterfaceTypes struct {
	types []InterfaceType
}

func (types InterfaceTypes) Has(t InterfaceType) bool {
	for i := 0; i < len(types.types); i++ {
		item := types.types[i]
		if item == t {
			return true
		}
	}

	return false
}

func (types InterfaceTypes) run(t InterfaceType) error {
	if types.Has(t) {
		return fmt.Errorf("already running")
	}

	return nil
}

type Interfaces struct {
	ctx            context.Context
	messageQueue   chan string
	Http_server    *bot_interface_http.Server
	Daemon         *bot_interface_daemon.Daemon // not ready but an example of an interface
	interfaceTypes *InterfaceTypes
}

func New(ctx context.Context) *Interfaces {
	it := &InterfaceTypes{
		types: make([]InterfaceType, TotalInterfaceTypes),
	}
	return &Interfaces{
		ctx:            ctx,
		interfaceTypes: it,
	}
}

func (b *Interfaces) Listen() {
	switch {
	case b.HasHttpServerRunning():
		b.Http_server.Listen()
	case b.HasDaemonRunning():
		b.Daemon.Listen()
	}
}

// TODO: This design is faulty:
//    it has the format of Init<InterfaceName>/Start<InterfaceName> (e.g. StartHttpServer()).
//    Obviously, not quite practical, what if a new interface pops ups, what if we haven't implemented some of necessary functions? (no interface usage).
//    And it's so much mis-usage of namespace,
//      normally used be bot.Interfaces.Http.Init/Start, not bot.Interfaces.StartHttpServer.
//    This should be fixed in the future but it's a fast way of implementing this now.

func (b *Interfaces) InitHttpServer(addr string, certFile *string, keyFile *string) error {
	if b.interfaceTypes.Has(InterfaceTypeHttpServer) {
		return fmt.Errorf("already running http server")
	}

	err := b.interfaceTypes.run(InterfaceTypeHttpServer)
	if err != nil {
		return err
	}

	if b.Http_server == nil {
		b.Http_server = bot_interface_http.New(bot_interface_http.Server_Args{
			Context:  b.ctx,
			Address:  addr,
			CertFile: certFile,
			KeyFile:  keyFile,
		})
	}

	return nil
}

func (b *Interfaces) StartHttpServer() error {
	return nil
}

func (b *Interfaces) HasHttpServerRunning() bool {
	return b.Http_server != nil && b.interfaceTypes.Has(InterfaceTypeHttpServer)
}

func (b *Interfaces) HasDaemonRunning() bool {
	return b.Daemon != nil && b.interfaceTypes.Has(InterfaceTypeDaemon)
}
