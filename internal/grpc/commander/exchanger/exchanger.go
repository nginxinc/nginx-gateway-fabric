package exchanger

import (
	"context"

	"github.com/nginx/agent/sdk/v2/proto"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 . CommandExchanger

// CommandExchanger exchanges commands between a client and server through the In() and Out() methods.
// Run runs the exchanger.
// To send a command to the client, place it on the Out() channel.
// To receive a command from the client, listen to the In() chanel.
type CommandExchanger interface {
	// Run the CommandExchanger.
	Run(ctx context.Context) error
	// Out returns a write-only channel of commands.
	// Commands placed on this channel are sent to the client.
	Out() chan<- *proto.Command
	// In returns a read-only channel of commands.
	// Commands read from this channel are from the client.
	In() <-chan *proto.Command
}
