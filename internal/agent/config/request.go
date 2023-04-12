package config

import (
	"context"

	"github.com/nginx/agent/sdk/v2/proto"
)

// Request is a request for a *proto.Config.
type Request struct {
	replyCh chan *proto.NginxConfig
	id      string
}

// NewRequest returns a Request with the provided ID.
func NewRequest(id string) *Request {
	return &Request{
		id:      id,
		replyCh: make(chan *proto.NginxConfig),
	}
}

// WaitForReply blocks until a reply is received or the context is canceled.
func (r *Request) WaitForReply(ctx context.Context) (*proto.NginxConfig, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case reply := <-r.replyCh:
		return reply, nil
	}
}

// reply replies to the request with the provided config. Blocks until reply is received or the context is canceled.
func (r *Request) reply(ctx context.Context, config *proto.NginxConfig) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case r.replyCh <- config:
		return nil
	}
}
