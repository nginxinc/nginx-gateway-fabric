package commander

import (
	"context"
	"errors"
	"fmt"

	"github.com/go-logr/logr"
	"github.com/nginx/agent/sdk/v2/proto"
	"golang.org/x/sync/errgroup"
)

const channelLength = 25

// BidirectionalChannel encapsulates the CommandChannelServer which is a bidirectional streaming channel.
// The BidirectionalChannel is responsible for sending and receiving commands to and from the CommandChannelServer.
//
// All commands received from the CommandChannelServer are put on the fromClient channel and can be accessed through
// the Out() method.
//
// Commands can be sent to the CommandChannelServer by placing them on the toClient channel,
// which is accessible through the In() method.
//
// To use the BidirectionalChannel you must call the Run() method to kick of the receive and send loops.
type BidirectionalChannel struct {
	channel    proto.Commander_CommandChannelServer
	fromClient chan *proto.Command
	toClient   chan *proto.Command
	logger     logr.Logger
}

// NewBidirectionalChannel returns a new instance of the BidirectionalChannel.
func NewBidirectionalChannel(
	channel proto.Commander_CommandChannelServer,
	logger logr.Logger,
) *BidirectionalChannel {
	return &BidirectionalChannel{
		channel:    channel,
		fromClient: make(chan *proto.Command, channelLength),
		toClient:   make(chan *proto.Command, channelLength),
		logger:     logger,
	}
}

// Run runs the receive and send loops on the BidirectionalChannel.
// Run is blocking and will return if an error occurs in either loop or the context is canceled.
func (bc *BidirectionalChannel) Run(parent context.Context) error {
	defer func() {
		close(bc.fromClient)
		close(bc.toClient)
	}()

	eg, ctx := errgroup.WithContext(parent)

	eg.Go(func() error {
		return bc.receive(ctx)
	})

	eg.Go(func() error {
		return bc.send(ctx)
	})

	return eg.Wait()
}

func (bc *BidirectionalChannel) receive(ctx context.Context) error {
	defer func() {
		bc.logger.Info("Stopping receive command loop")
	}()

	bc.logger.Info("Starting receive command loop")

	for {
		cmd, err := bc.channel.Recv()
		if err != nil {
			return fmt.Errorf("error receiving command from CommandChannel: %w", err)
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			if cmd != nil {
				bc.logger.Info("Received command", "command type", fmt.Sprintf("%T", cmd.Data))

				select {
				case <-ctx.Done():
					return ctx.Err()
				case bc.fromClient <- cmd:
				}
			} else {
				// The agent should never send us a nil command, but we catch this case out of an abundance of caution.
				// We don't want to return an error in this case because that would break the CommandChannel
				// connection with the agent. Instead, we log the abnormality and continue processing.
				bc.logger.Error(errors.New("received nil command"), "expected non-nil command")
			}
		}
	}
}

func (bc *BidirectionalChannel) send(ctx context.Context) error {
	defer func() {
		bc.logger.Info("Stopping send command loop")
	}()

	bc.logger.Info("Starting send command loop")

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case cmd := <-bc.toClient:
			if cmd == nil {
				panic("outgoing command is nil")
			}
			bc.logger.Info("Sending command", "command", cmd)
			if err := bc.channel.Send(cmd); err != nil {
				return fmt.Errorf("error sending command to CommandChannel: %w", err)
			}
		}
	}
}

// In returns a write-only channel of commands.
// Commands written to this channel will be sent to the client over the CommandChannelServer.
func (bc *BidirectionalChannel) In() chan<- *proto.Command {
	return bc.toClient
}

// Out returns a read-only channel of commands.
// The BidirectionalChannel writes commands that it receives from the CommandChannelServer to this channel.
func (bc *BidirectionalChannel) Out() <-chan *proto.Command {
	return bc.fromClient
}
