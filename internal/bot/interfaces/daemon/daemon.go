package bot_interface_daemon

import (
	"context"
	"os"
)

type Daemon interface {
	Start() <-chan error
	Listen() <-chan []byte
}

type daemon struct {
	ctx  context.Context
	File *os.File
}

type NewAndRunArgs struct {
	Context context.Context
}

func NewAndRun(args NewAndRunArgs) Daemon {
	daemon := &daemon{
		ctx: args.Context,
	}
	return daemon
}

func (d *daemon) Start() <-chan error {
	return nil
}

func (d *daemon) Listen() <-chan []byte {
	return nil
}
