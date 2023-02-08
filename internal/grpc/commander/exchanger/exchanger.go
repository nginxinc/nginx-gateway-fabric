package exchanger

import (
	"context"

	"github.com/nginx/agent/sdk/v2/proto"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 . CommandExchanger

// CommandExchanger exchanges commands between two components through the In() and Out() methods.
// Run runs the exchanger.
// To send a command, place it on the In() channel.
// To receive a command, listen to the Out() chanel.
type CommandExchanger interface {
	// Run the CommandExchanger.
	Run(ctx context.Context) error
	// In returns a write-only channel of commands.
	In() chan<- *proto.Command
	// Out returns a read-only channel of commands.
	Out() <-chan *proto.Command
}
