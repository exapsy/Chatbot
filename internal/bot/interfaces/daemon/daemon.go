package bot_interface_daemon

import (
	"context"
	"os"
)

type Daemon struct {
	ctx  context.Context
	File *os.File
}

type NewAndRunArgs struct {
	Context context.Context
}

func NewAndRun(args NewAndRunArgs) *Daemon {
	daemon := &Daemon{
		ctx: args.Context,
	}
	return daemon
}

func (d *Daemon) Start() <-chan error {
	return nil
}

func (d *Daemon) Listen() <-chan string {
	return nil
}
