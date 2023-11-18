package bot_interfaces

import (
	"connectly-interview/internal/bot/domain/bot_chat"
	"connectly-interview/internal/bot/interfaces/daemon"
	"connectly-interview/internal/bot/interfaces/http_server"
	"context"
	"fmt"
)

// TODO: refactoring - this can be rethought and just have an array of communication interfaces
//        instead of having each interface separately and separate functions inside here for each one.
//        Basically ... Fast non-easily-scaleable architectural decision was taken here for the sake of doing it fast.

// Interface is the interface that describes
// which functions and behaviors a communication interface should be compatible with.
type Interface interface {
	// Start starts the communication interface
	Start() <-chan error
	// Listen listens to the prompts of the interface
	Listen() <-chan []byte
}

// TotalInterfaceTypes keeps track of how many interfaces there are
const TotalInterfaceTypes = 3

// TODO: Dangerous: it does not auto-update/generate // if the amount of interfaces changes, REMEMBER to change this
//	 I don't know how, implement this in a different way. Unfortunately Golang does not have Enums, cant reflect on enums coz they dont exist ...
//	 so it's the best way for now, but let's rethink this in the future shall we?

// InterfaceType describes which communication interface it is (e.g. HTTP, grpc etc.)
type InterfaceType uint8

const (
	InterfaceTypeHttpServer = iota + 1
	InterfaceTypeDaemon     // To be implemented
	InterfaceTypeGrpc       // To be implemented
)

type InterfaceTypes struct {
	types []InterfaceType
}

// HasAny returns if there are any running communication interface types
func (types InterfaceTypes) HasAny() bool {
	for i := 0; i < len(types.types); i++ {
		item := types.types[i]
		if item != 0 {
			return true
		}
	}

	return false
}

// Has returns if the communication interface is running
func (types InterfaceTypes) Has(t InterfaceType) bool {
	for i := 0; i < len(types.types); i++ {
		item := types.types[i]
		if item == t {
			return true
		}
	}

	return false
}

// run checks if the interface is already running and if not it adds it to the list of running interfaces
// so we can know which communication interfaces are running.
func (types InterfaceTypes) run(t InterfaceType) error {
	if types.Has(t) {
		return fmt.Errorf("already running")
	}

	// interfaceTypes start from iota+1 = 1,
	// and `types` is of length equal to the total interfaces,
	// so we reduce it by 1
	index := int(t) - 1
	types.types[index] = t

	return nil
}

type Interfaces struct {
	ctx context.Context
	// messageQueue is where the channel where all the messages of all the communication interfaces will come through
	messageQueue chan []byte
	// Http_server is literally the http server that will act as another communication interface
	Http_server bot_interface_http.Server
	// Daemon is a daemon socket communication interface
	Daemon bot_interface_daemon.Daemon // not ready but an example of an interface
	// interfaceTypes keeps which interface types are running
	interfaceTypes        *InterfaceTypes
	newChatHandler        func()
	newChatMessageHandler func(chatId bot_chat.ChatId, msg []byte) error
}

type Option func(i *Interfaces)

func WithMessageQueueCapacity(cap uint32) Option {
	return func(i *Interfaces) {
		i.messageQueue = make(chan []byte, cap)
	}
}

func WithNewChatHandler(cb func()) Option {
	return func(i *Interfaces) {
		i.newChatHandler = cb
	}
}

func WithNewChatMessageHandler(cb func(chatId bot_chat.ChatId, msg []byte) error) Option {
	return func(i *Interfaces) {
		i.newChatMessageHandler = cb
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
		interfaces.messageQueue = make(chan []byte, 4)
	}

	return interfaces
}

// Listen listens to all the interfaces that let the outside world communicate with the bot
// and writes the prompts that come to the interface's message queue
func (b *Interfaces) Listen() (prompts <-chan []byte) {
	if b.interfaceTypes.HasAny() {
		fmt.Printf("listening to communication interfaces:\n")
	}

	if b.HasHttpServerRunning() {
		promptChan, errChan := b.Http_server.Listen()
		fmt.Printf("\t- http server\n")
		go func() {
			for {
				select {
				case prompt := <-promptChan:
					b.messageQueue <- prompt
				case err := <-errChan:
					fmt.Printf("error on http server request: %s", err)
				}
			}
		}()
	}

	if b.HasDaemonRunning() {
		promptChan := b.Daemon.Listen()
		fmt.Printf("\t - daemon socket\n")
		go func() {
			for {
				select {
				case prompt := <-promptChan:
					b.messageQueue <- prompt
				}
			}
		}()
	}

	return b.messageQueue
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
			Context:               b.ctx,
			Address:               addr,
			CertFile:              certFile,
			KeyFile:               keyFile,
			ExtraMiddlewares:      nil,
			NewChatHandler:        b.newChatHandler,
			NewChatMessageHandler: b.newChatMessageHandler,
		})
	}

	fmt.Println("initiated http server")

	return nil
}

func (b *Interfaces) HasHttpServerRunning() bool {
	return b.Http_server != nil && b.interfaceTypes.Has(InterfaceTypeHttpServer)
}

func (b *Interfaces) HasDaemonRunning() bool {
	return b.Daemon != nil && b.interfaceTypes.Has(InterfaceTypeDaemon)
}
