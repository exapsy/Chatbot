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
	Listen() <-chan []byte
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

func (types InterfaceTypes) HasAny() bool {
	for i := 0; i < len(types.types); i++ {
		item := types.types[i]
		if item != 0 {
			return true
		}
	}

	return false
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
	Http_server    bot_interface_http.Server
	Daemon         bot_interface_daemon.Daemon // not ready but an example of an interface
	interfaceTypes *InterfaceTypes
}

type Option func(i *Interfaces)

func WithMessageQueueCapacity(cap uint32) Option {
	return func(i *Interfaces) {
		i.messageQueue = make(chan string, cap)
	}
}

func New(ctx context.Context, opts ...Option) *Interfaces {
	it := &InterfaceTypes{
		types: make([]InterfaceType, TotalInterfaceTypes),
	}

	interfaces := &Interfaces{
		ctx:            ctx,
		interfaceTypes: it,
	}

	for _, o := range opts {
		o(interfaces)
	}

	if interfaces.messageQueue == nil {
		interfaces.messageQueue = make(chan string, 4)
	}

	return interfaces
}

// Listen listens to all the interfaces that let the outside world communicate with the bot
// and writes the prompts that come to the interface's message queue
func (b *Interfaces) Listen() (prompts <-chan []byte) {
	if b.HasHttpServerRunning() {
		promptChan := b.Http_server.Listen()
		fmt.Printf("listening to http server ...\n")
		go func() {
			for {
				select {
				case prompt := <-promptChan:
					b.messageQueue <- string(prompt)
				}
			}
		}()
	}

	if b.HasDaemonRunning() {
		promptChan := b.Daemon.Listen()
		fmt.Printf("listening to daemon socket ...\n")
		go func() {
			for {
				select {
				case prompt := <-promptChan:
					b.messageQueue <- string(prompt)
				}
			}
		}()
	}

	return prompts
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
	b.Http_server.Start()
	return nil
}

func (b *Interfaces) HasHttpServerRunning() bool {
	return b.Http_server != nil && b.interfaceTypes.Has(InterfaceTypeHttpServer)
}

func (b *Interfaces) HasDaemonRunning() bool {
	return b.Daemon != nil && b.interfaceTypes.Has(InterfaceTypeDaemon)
}
